package commands

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/yhlooo/kubectl-cache/pkg/commands/options"
	"github.com/yhlooo/kubectl-cache/pkg/proxymgr"
)

// NewShutdownCommandWithOptions 使用指定选项创建 shutdown 子命令
func NewShutdownCommandWithOptions(opts *options.ShutdownOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shutdown [NAME...]",
		Short: "Shutdown the cache proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx)
			globalOpts := options.GlobalOptionsFromContext(ctx)
			mgr := proxymgr.NewProxyManager(globalOpts.DataRoot, nil)

			// 获取需要关闭的代理列表
			var proxies []proxymgr.Proxy
			if opts.All {
				proxyList, err := mgr.List(ctx)
				if err != nil {
					return fmt.Errorf("list all proxies error: %w", err)
				}
				proxies = proxyList.Items
			} else {
				if len(args) == 0 {
					return fmt.Errorf("no proxy names specified")
				}
				for _, name := range args {
					p, err := mgr.Get(ctx, name)
					if err != nil {
						logger.Info(fmt.Sprintf("WARNING get proxy %q error: %v", name, err))
						continue
					}
					proxies = append(proxies, *p)
				}
			}

			for _, p := range proxies {
				if err := mgr.KillProxy(ctx, &p, opts.Wait, opts.Force); err != nil {
					logger.Info(fmt.Sprintf("WARNING kill proxy %q error: %v", p.Name, err))
				}
			}

			return nil
		},
	}

	// 绑定选项到命令行参数
	opts.AddPFlags(cmd.Flags())

	return cmd
}
