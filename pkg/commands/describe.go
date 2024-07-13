package commands

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmddescribe "k8s.io/kubectl/pkg/cmd/describe"
	utilcomp "k8s.io/kubectl/pkg/util/completion"

	"github.com/yhlooo/kubectl-cache/pkg/utils/cmdutil"
)

// NewDescribeCommand 创建 describe 子命令
func NewDescribeCommand(clientGetter genericclioptions.RESTClientGetter) *cobra.Command {
	return cmdutil.NewKubectlCommandWithInternalProxy(
		clientGetter,
		"cache",
		cmddescribe.NewCmdDescribe,
		utilcomp.ResourceTypeAndNameCompletionFunc,
	)
}
