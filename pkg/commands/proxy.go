package commands

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	kubectlproxy "k8s.io/kubectl/pkg/proxy"

	"github.com/yhlooo/kubectl-cache/pkg/commands/options"
	"github.com/yhlooo/kubectl-cache/pkg/proxy"
)

// NewProxyCommandWithOptions 基于选项创建 proxy 子命令
func NewProxyCommandWithOptions(opts *options.ProxyOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Run a proxy to the Kubernetes API server",
		Long: `Creates a proxy server or application-level gateway between localhost and the Kubernetes API server. It also allows
serving static content over specified HTTP path. All incoming data enters through one port and gets forwarded to the
remote Kubernetes API server port, except for the path matching the static content path.`,
		Example: `  # To proxy all of the Kubernetes API and nothing else
  kubectl proxy --api-prefix=/

  # To proxy only part of the Kubernetes API and also some static files
  # You can get pods info with 'curl localhost:8001/api/v1/pods'
  kubectl proxy --www=/my/files --www-prefix=/static/ --api-prefix=/api/

  # To proxy the entire Kubernetes API at a different root
  # You can get pods info with 'curl localhost:8001/custom/api/v1/pods'
  kubectl proxy --api-prefix=/custom/

  # Run a proxy to the Kubernetes API server on port 8011, serving static content from ./local/www/
  kubectl proxy --port=8011 --www=./local/www/

  # Run a proxy to the Kubernetes API server on an arbitrary local port
  # The chosen port for the server will be output to stdout
  kubectl proxy --port=0

  # Run a proxy to the Kubernetes API server, changing the API prefix to k8s-api
  # This makes e.g. the pods API available at localhost:8001/k8s-api/v1/pods/
  kubectl proxy --api-prefix=/k8s-api`,

		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logr.FromContextOrDiscard(ctx)

			// 加载 Kubernetes 客户端配置
			globalOpts := options.GlobalOptionsFromContext(ctx)
			config, err := globalOpts.ClientConfig.ToRESTConfig()
			if err != nil {
				return err
			}
			mapper, err := globalOpts.ClientConfig.ToRESTMapper()
			if err != nil {
				return err
			}

			// 处理过滤选项
			var filter *kubectlproxy.FilterServer
			if opts.DisableFilter && opts.UNIXSocket == "" {
				logger.Info("WARNING request filter disabled, " +
					"your proxy is vulnerable to XSRF attacks, please be cautious")
			} else {
				filter = &kubectlproxy.FilterServer{
					AcceptPaths:   kubectlproxy.MakeRegexpArrayOrDie(opts.AcceptPaths),
					RejectPaths:   kubectlproxy.MakeRegexpArrayOrDie(opts.RejectPaths),
					AcceptHosts:   kubectlproxy.MakeRegexpArrayOrDie(opts.AcceptHosts),
					RejectMethods: kubectlproxy.MakeRegexpArrayOrDie(opts.RejectMethods),
				}
			}

			// 处理监听选项
			listenerOpts := proxy.ListenerOptions{}
			if opts.UNIXSocket != "" {
				listenerOpts.UNIXSocket = &proxy.UNIXSocketListenerOptions{
					Path: opts.UNIXSocket,
				}
			} else {
				listenerOpts.TCP = &proxy.TCPListenerOptions{
					Address: opts.Address,
					Port:    opts.Port,
				}
			}

			// 创建代理服务
			s, err := proxy.NewServer(ctx, proxy.ServerOptions{
				ClientConfig: config,
				RESTMapper:   mapper,
				Listener:     listenerOpts,
				APIProxy: proxy.APIProxyServerOptions{
					URIPrefix:          opts.APIPrefix,
					Filter:             filter,
					Keepalive:          opts.Keepalive,
					AppendLocationPath: opts.AppendServerPath,
				},
				Static: proxy.StaticServerOptions{
					URIPrefix: opts.WWWPrefix,
					FileBase:  opts.WWW,
				},
			})
			if err != nil {
				return fmt.Errorf("create proxy server error: %w", err)
			}

			// 启动代理服务
			return s.Serve(ctx)
		},
	}

	// 绑定命令行参数到选项
	opts.AddPFlags(cmd.Flags())

	return cmd
}
