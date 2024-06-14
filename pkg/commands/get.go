package commands

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cmdget "k8s.io/kubectl/pkg/cmd/get"
	kubectlcmdutil "k8s.io/kubectl/pkg/cmd/util"
	utilcomp "k8s.io/kubectl/pkg/util/completion"

	"github.com/yhlooo/kubectl-cache/pkg/commands/options"
	"github.com/yhlooo/kubectl-cache/pkg/proxyclientgetter"
	"github.com/yhlooo/kubectl-cache/pkg/proxymgr"
)

// NewGetCommand 创建 get 子命令
func NewGetCommand(clientGetter genericclioptions.RESTClientGetter) *cobra.Command {
	proxyClientGetter := &proxyclientgetter.ProxyClientGetter{
		RESTClientGetter: clientGetter,
	}
	f := kubectlcmdutil.NewFactory(proxyClientGetter)
	cmd := cmdget.NewCmdGet("cache", f, genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})
	cmd.ValidArgsFunction = utilcomp.ResourceTypeAndNameCompletionFunc(f)

	// 修改默认 PreRun 方法
	oldPreRunE := cmd.PreRunE
	oldPreRun := cmd.PreRun
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		globalOpts := options.GlobalOptionsFromContext(ctx)

		// 注入上下文和代理管理器
		proxyClientGetter.SetContext(ctx)
		proxyClientGetter.ProxyManager = proxymgr.NewProxyManager(globalOpts.DataRoot, GetStartInternalProxyArgs(cmd))

		if oldPreRunE != nil {
			return oldPreRunE(cmd, args)
		}
		if oldPreRun != nil {
			oldPreRun(cmd, args)
		}
		return nil
	}

	return cmd
}
