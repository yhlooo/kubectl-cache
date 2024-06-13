package commands

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/kubectl-cache/pkg/commands/options"
	"github.com/yhlooo/kubectl-cache/pkg/proxy"
	"github.com/yhlooo/kubectl-cache/pkg/proxymgr"
)

// NewInternalProxyCommandWithOptions 基于选项创建一个 internal-proxy 子命令
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

			mgr := proxymgr.NewProxyManager(globalOpts.DataRoot)

			// 锁
			info, err := mgr.LockConfig(ctx, config)
			if err != nil {
				return fmt.Errorf("get lock for client config error: %w", err)
			}
			defer func() {
				if err := mgr.UnlockConfig(ctx, config); err != nil {
					logger.Error(err, "unlock client config error")
				}
			}()

			listenerOpts := proxy.ListenerOptions{}
			switch runtime.GOOS {
			case "darwin", "linux":
				listenerOpts.UNIXSocket = &proxy.UNIXSocketListenerOptions{
					Path: info.UNIXSocketPath,
				}
			default:
				listenerOpts.TCP = &proxy.TCPListenerOptions{
					Address: "127.0.0.1",
					Port:    info.Port,
				}
			}

			// 创建代理服务
			s, err := proxy.NewServer(ctx, proxy.ServerOptions{
				ClientConfig: config,
				RESTMapper:   mapper,
				Listener:     listenerOpts,
				APIProxy: proxy.APIProxyServerOptions{
					URIPrefix: "/",
				},
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
				return ctx.Err()
			case <-s.Ready():
			}

			// 更新下记录的代理端口信息
			if listenerOpts.TCP != nil {
				addr := s.Addr().String()
				dividedAddr := strings.SplitN(addr, ":", 2)
				if len(dividedAddr) == 0 {
					return fmt.Errorf("invalid address: %s", addr)
				}
				port, err := strconv.ParseInt(dividedAddr[1], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid address: %s", addr)
				}

				info.UNIXSocketPath = ""
				info.Port = int(port)
				if err := mgr.SetProxyInfoForConfig(ctx, config, info); err != nil {
					return fmt.Errorf("set proxy info error: %w", err)
				}
			}

			// 等待代理服务结束
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-done:
			}

			return serveErr
		},
	}

	// 绑定命令行参数到选项
	opts.AddPFlags(cmd.Flags())

	return cmd
}
