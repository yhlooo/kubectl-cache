package commands

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cmdget "k8s.io/kubectl/pkg/cmd/get"
	kubectlcmdutil "k8s.io/kubectl/pkg/cmd/util"
	utilcomp "k8s.io/kubectl/pkg/util/completion"
)

// NewGetCommand 创建 get 子命令
func NewGetCommand(f kubectlcmdutil.Factory) *cobra.Command {
	cmd := cmdget.NewCmdGet("cache", f, genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})
	cmd.ValidArgsFunction = utilcomp.ResourceTypeAndNameCompletionFunc(f)
	return cmd
}
