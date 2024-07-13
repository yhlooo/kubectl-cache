package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	// 注册认证插件
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/yhlooo/kubectl-cache/pkg/commands/options"
	"github.com/yhlooo/kubectl-cache/pkg/utils/cmdutil"
)

// NewRootCommand 创建一个 kubectl-cache 命令
func NewRootCommand() *cobra.Command {
	return NewRootCommandWithOptions(options.NewDefaultOptions())
}

// NewRootCommandWithOptions 基于选项创建一个 kubectl-cache 命令
func NewRootCommandWithOptions(opts options.Options) *cobra.Command {
	displayName := "kubectl-cache"
	if lastArg := os.Getenv("_"); lastArg != "" {
		switch filepath.Base(lastArg) {
		case "kubectl", "kubectl.exe":
			// 当前是以 kubectl 插件方式运行的
			displayName = "kubectl cache"
		}
	}
	cmd := &cobra.Command{
		Use: displayName,
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: displayName,
		},
		Short:        "Get or list Kubernetes resources with local cache",
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
		NewGetCommand(opts.Global.ClientConfig),
		NewDescribeCommand(opts.Global.ClientConfig),
		NewProxyCommandWithOptions(&opts.Proxy),
		NewProxiesCommandWithOptions(&opts.Proxies),
		NewShutdownCommandWithOptions(&opts.Shutdown),
		NewVersionCommandWithOptions(&opts.Version),
		NewInternalProxyCommandWithOptions(&opts.InternalProxyOptions),
	)

	return cmd
}
