package commands

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmddescribe "k8s.io/kubectl/pkg/cmd/describe"

	"github.com/yhlooo/kubectl-cache/pkg/utils/cmdutil"
)

// NewDescribeCommand 创建 describe 子命令
func NewDescribeCommand(clientGetter genericclioptions.RESTClientGetter) *cobra.Command {
	_, cmd := cmdutil.NewKubectlCommandWithInternalProxy(clientGetter, "cache", cmddescribe.NewCmdDescribe)
	return cmd
}
