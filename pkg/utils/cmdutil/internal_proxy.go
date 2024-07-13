package cmdutil

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	kubectlcmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/yhlooo/kubectl-cache/pkg/commands/options"
	"github.com/yhlooo/kubectl-cache/pkg/proxyclientgetter"
	"github.com/yhlooo/kubectl-cache/pkg/proxymgr"
)

// NewKubectlCommandFunc 创建 kubectl 命令方法
type NewKubectlCommandFunc func(
	parent string,
	f kubectlcmdutil.Factory,
	streams genericiooptions.IOStreams,
) *cobra.Command

// NewValidArgsFunc 生成 cobra.Command.ValidArgsFunction 方法
type NewValidArgsFunc func(
	f kubectlcmdutil.Factory,
) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

// NewKubectlCommandWithInternalProxy 创建使用缓存代理的 kubectl 命令
func NewKubectlCommandWithInternalProxy(
	clientGetter genericclioptions.RESTClientGetter,
	parent string,
	newCmd NewKubectlCommandFunc,
	newValidArgsFunc NewValidArgsFunc,
) *cobra.Command {
	proxyClientGetter := &proxyclientgetter.ProxyClientGetter{
		RESTClientGetter: clientGetter,
	}
	f := kubectlcmdutil.NewFactory(proxyClientGetter)

	// 创建命令
	cmd := newCmd(parent, f, genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})

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

	// 修改 ValidArgsFunction 方法
	if newValidArgsFunc != nil {
		validArgsFunction := newValidArgsFunc(f)
		cmd.ValidArgsFunction = func(
			cmd *cobra.Command,
			args []string,
			toComplete string,
		) ([]string, cobra.ShellCompDirective) {
			ctx := cmd.Context()
			globalOpts := options.GlobalOptionsFromContext(ctx)

			// 注入上下文和代理管理器
			proxyClientGetter.SetContext(ctx)
			proxyClientGetter.ProxyManager = proxymgr.NewProxyManager(
				globalOpts.DataRoot,
				GetStartInternalProxyArgs(cmd),
			)

			return validArgsFunction(cmd, args, toComplete)
		}
	}

	return cmd
}

// GetStartInternalProxyArgs 获取启动代理服务的命令行参数
func GetStartInternalProxyArgs(cmd *cobra.Command) []string {
	args := []string{"internal-proxy"}

	globalOpts := options.GlobalOptionsFromContext(cmd.Context())
	if globalOpts.ClientConfig == nil {
		return args
	}
	clientConfig := globalOpts.ClientConfig

	for k, v := range map[string]*string{
		"--kubeconfig":            clientConfig.KubeConfig,
		"--cache-dir":             clientConfig.CacheDir,
		"--client-certificate":    clientConfig.CertFile,
		"--client-key":            clientConfig.KeyFile,
		"--token":                 clientConfig.BearerToken,
		"--as":                    clientConfig.Impersonate,
		"--as-uid":                clientConfig.ImpersonateUID,
		"--username":              clientConfig.Username,
		"--password":              clientConfig.Password,
		"--cluster":               clientConfig.ClusterName,
		"--user":                  clientConfig.AuthInfoName,
		"--namespace":             clientConfig.Namespace,
		"--context":               clientConfig.Context,
		"--server":                clientConfig.APIServer,
		"--tls-server-name":       clientConfig.TLSServerName,
		"--certificate-authority": clientConfig.CAFile,
		"--request-timeout":       clientConfig.Timeout,
	} {
		if v != nil && *v != "" {
			args = append(args, k, *v)
		}
	}
	if clientConfig.ImpersonateGroup != nil {
		for _, g := range *clientConfig.ImpersonateGroup {
			args = append(args, "--as-group", g)
		}
	}
	if clientConfig.Insecure != nil && *clientConfig.Insecure {
		args = append(args, "--insecure-skip-tls-verify")
	}
	if clientConfig.DisableCompression != nil && *clientConfig.DisableCompression {
		args = append(args, "--disable-compression")
	}

	return args
}
