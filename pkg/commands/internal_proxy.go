package commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/kubectl-cache/pkg/commands/options"
	"github.com/yhlooo/kubectl-cache/pkg/proxy"
	"github.com/yhlooo/kubectl-cache/pkg/proxymgr"
)

// NewInternalProxyCommandWithOptions 基于选项创建 internal-proxy 子命令
// 该命令和 proxy 子命令作用基本一样，只不过专用于实现内部逻辑（比如 get 命令）
func NewInternalProxyCommandWithOptions(opts *options.InternalProxyOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "internal-proxy",
		Short:  "Run a proxy to the Kubernetes API server (internal component, DO NOT USE)",
		Hidden: true, // 隐藏内部命令
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx)
			globalOpts := options.GlobalOptionsFromContext(ctx)

			// 加载 Kubernetes 客户端配置
			config, err := globalOpts.ClientConfig.ToRESTConfig()
			if err != nil {
				return err
			}
			mapper, err := globalOpts.ClientConfig.ToRESTMapper()
			if err != nil {
				return err
			}

			mgr := proxymgr.NewProxyManager(globalOpts.DataRoot, nil)

			// 锁
			proxyObj, err := mgr.LockProxy(ctx, config)
			if err != nil {
				return fmt.Errorf("get lock for client config error: %w", err)
			}
			defer func() {
				if err := mgr.UnlockProxy(ctx, proxyObj); err != nil {
					logger.Error(err, "unlock client config error")
				}
			}()

			// 创建代理服务
			s, err := proxy.NewServer(ctx, proxy.ServerOptions{
				ClientConfig: config,
				RESTMapper:   mapper,
				Listener: proxy.ListenerOptions{
					TCP: &proxy.TCPListenerOptions{
						Address: "127.0.0.1",
						Port:    0,
					},
				},
				APIProxy: proxy.APIProxyServerOptions{
					URIPrefix: "/",
				},
				MaxIdleTime: opts.MaxIdleTime,
			})
			if err != nil {
				return fmt.Errorf("create proxy server error: %w", err)
			}

			// 启动代理服务
			done := make(chan struct{})
			var serveErr error
			go func() {
				serveErr = s.Serve(ctx)
				close(done)
			}()

			// 等待代理服务就绪
			select {
			case <-ctx.Done():
				return fmt.Errorf("wait for proxy ready error: %w", ctx.Err())
			case <-s.Ready():
			}
			if serveErr != nil {
				return serveErr
			}

			// 更新下记录的代理端口信息
			addr := s.Addr().String()
			dividedAddr := strings.SplitN(addr, ":", 2)
			if len(dividedAddr) == 0 {
				return fmt.Errorf("invalid address: %s", addr)
			}
			port, err := strconv.ParseInt(dividedAddr[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid address: %s", addr)
			}
			proxyObj.Status.Port = int(port)
			if err := mgr.SetProxy(ctx, proxyObj); err != nil {
				return fmt.Errorf("set proxy info error: %w", err)
			}

			// 等待代理服务结束
			<-done

			if !errors.Is(serveErr, context.Canceled) {
				return serveErr
			}
			return nil
		},
	}

	// 绑定命令行参数到选项
	opts.AddPFlags(cmd.Flags())

	return cmd
}
