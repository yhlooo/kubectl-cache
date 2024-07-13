package commands

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdget "k8s.io/kubectl/pkg/cmd/get"
	utilcomp "k8s.io/kubectl/pkg/util/completion"

	"github.com/yhlooo/kubectl-cache/pkg/utils/cmdutil"
)

// NewGetCommand 创建 get 子命令
func NewGetCommand(clientGetter genericclioptions.RESTClientGetter) *cobra.Command {
	f, cmd := cmdutil.NewKubectlCommandWithInternalProxy(clientGetter, "cache", cmdget.NewCmdGet)
	cmd.ValidArgsFunction = utilcomp.ResourceTypeAndNameCompletionFunc(f)
	return cmd
}
