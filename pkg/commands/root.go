package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yhlooo/kubectl-cache/pkg/commands/options"
	"github.com/yhlooo/kubectl-cache/pkg/utils/cmdutil"
)

// NewRootCommand 创建一个 kubectl-cache 命令
func NewRootCommand() *cobra.Command {
	return NewRootCommandWithOptions(options.NewDefaultOptions())
}

// NewRootCommandWithOptions  使用指定选项创建一个 kubectl-cache 命令
func NewRootCommandWithOptions(opts options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kubectl-cache",
		Short:        "Get or List Kubernetes resources with local cache",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 校验全局选项
			if err := opts.Global.Validate(); err != nil {
				return err
			}
			// 设置日志
			logger := cmdutil.SetLogger(cmd, opts.Global.Verbosity)
			// 将全局选项设置到上下文
			cmd.SetContext(options.NewContextWithGlobalOptions(cmd.Context(), opts.Global))

			logger.V(1).Info(fmt.Sprintf("command: %q, args: %#v, options: %#v", cmd.Name(), args, opts))
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// 绑定选项到命令行参数
	opts.Global.AddPFlags(cmd.PersistentFlags())

	// 添加子命令
	cmd.AddCommand(
		NewProxyCommandWithOptions(&opts.Proxy),
		NewInternalProxyCommandWithOptions(&opts.InternalProxyOptions),
	)

	return cmd
}
