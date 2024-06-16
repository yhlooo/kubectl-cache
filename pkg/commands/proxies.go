package commands

import (
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
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
			logger := logr.FromContextOrDiscard(ctx)
			globalOpts := options.GlobalOptionsFromContext(ctx)
			mgr := proxymgr.NewProxyManager(globalOpts.DataRoot, nil)

			var ret runtime.Object
			var err error
			if len(args) == 0 {
				// 列出所有代理
				ret, err = mgr.List(ctx)
			} else if len(args) == 1 {
				// 获取指定一个代理
				ret, err = mgr.Get(ctx, args[0])
			} else {
				// 获取指定几个代理
				proxies := proxymgr.NewProxyList()
				ret = proxies
				for _, name := range args {
					p, err := mgr.Get(ctx, name)
					if err != nil {
						logger.Info(fmt.Sprintf("get proxy %q error: %v", name, err))
						continue
					}
					proxies.Items = append(proxies.Items, *p)
				}
			}
			if err != nil {
				return err
			}

			// 输出
			switch opts.OutputFormat {
			case "json":
				return (&printers.JSONPrinter{}).PrintObj(ret, os.Stdout)
			case "yaml":
				return (&printers.YAMLPrinter{}).PrintObj(ret, os.Stdout)
			default:
				table, err := proxymgr.ProxyTableConvertor.ConvertToTable(ctx, ret, nil)
				if err != nil {
					return fmt.Errorf("convert result to table error: %w", err)
				}
				return printers.NewTablePrinter(printers.PrintOptions{
					Wide:         false,
					ColumnLabels: nil,
					SortBy:       "",
				}).PrintObj(table, os.Stdout)
			}
		},
	}

	// 绑定选项到命令行参数
	opts.AddPFlags(cmd.Flags())

	return cmd
}
