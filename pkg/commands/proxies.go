package commands

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/printers"

	"github.com/yhlooo/kubectl-cache/pkg/commands/options"
	"github.com/yhlooo/kubectl-cache/pkg/proxymgr"
)

// NewProxiesCommandWithOptions 基于选项创建一个 proxies 子命令
func NewProxiesCommandWithOptions(opts *options.ProxiesOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxies",
		Short: "List cache proxies",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// 校验选项
			if err := opts.Validate(); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			globalOpts := options.GlobalOptionsFromContext(ctx)
			mgr := proxymgr.NewProxyManager(globalOpts.DataRoot, nil)

			// 列出所有代理
			proxies, err := mgr.List(ctx)
			if err != nil {
				return err
			}

			// 输出
			switch opts.OutputFormat {
			case "json":
				return (&printers.JSONPrinter{}).PrintObj(proxies, os.Stdout)
			case "yaml":
				return (&printers.YAMLPrinter{}).PrintObj(proxies, os.Stdout)
			default:
				return printers.NewTablePrinter(printers.PrintOptions{
					Wide:         false,
					ColumnLabels: nil,
					SortBy:       "",
				}).PrintObj(proxies, os.Stdout)
			}
		},
	}

	// 绑定选项到命令行参数
	opts.AddPFlags(cmd.Flags())

	return cmd
}
