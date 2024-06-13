package options

import (
	"time"

	"github.com/spf13/pflag"
)

// NewDefaultProxyOptions 创建一个默认的 proxy 子命令选项
func NewDefaultProxyOptions() ProxyOptions {
	return ProxyOptions{
		Address:    "127.0.0.1",
		Port:       8001,
		UNIXSocket: "",

		AcceptHosts:   `^localhost$,^127\.0\.0\.1$,^\[::1\]$`,
		AcceptPaths:   `^.*`,
		RejectMethods: `^$`,
		RejectPaths:   `^/api/.*/pods/.*/exec,^/api/.*/pods/.*/attach`,
		DisableFilter: false,

		APIPrefix:        "/",
		AppendServerPath: false,
		WWW:              "",
		WWWPrefix:        "/static/",
		Keepalive:        0,
	}
}

// ProxyOptions proxy 子命令选项
type ProxyOptions struct {
	// 监听地址
	Address string
	// 监听端口
	Port int
	// 监听 UNIX Socket
	UNIXSocket string

	// 允许访问的 Host
	AcceptHosts string
	// 允许访问的 URI 路径
	AcceptPaths string
	// 禁止访问的方法
	RejectMethods string
	// 禁止访问的 URI 路径
	RejectPaths string
	// 禁用过滤器
	DisableFilter bool

	// Kubernetes API URI 前缀
	APIPrefix string
	// 添加服务路径
	AppendServerPath bool
	// 静态文件路径
	WWW string
	// 静态文件服务 URI 前缀
	WWWPrefix string
	// 连接保持时长
	Keepalive time.Duration
}

// AddPFlags 将选项绑定到命令行参数
func (o *ProxyOptions) AddPFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.AcceptHosts, "accept-hosts", o.AcceptHosts, "Regular expression for hosts that the proxy should accept.")
	flags.StringVar(&o.AcceptPaths, "accept-paths", o.AcceptPaths, "Regular expression for paths that the proxy should accept.")
	flags.StringVar(&o.Address, "address", o.Address, "The IP address on which to serve on.")
	flags.StringVar(&o.APIPrefix, "api-prefix", o.APIPrefix, "Prefix to serve the proxied API under.")
	flags.BoolVar(&o.AppendServerPath, "append-server-path", o.AppendServerPath, "If true, enables automatic path appending of the kube context server path to each request.")
	flags.BoolVar(&o.DisableFilter, "disable-filter", o.DisableFilter, "If true, disable request filtering in the proxy. This is dangerous, and can leave you vulnerable to XSRF attacks, when used with an accessible port.")
	flags.DurationVar(&o.Keepalive, "keepalive", o.Keepalive, "keepalive specifies the keep-alive period for an active network connection. Set to 0 to disable keepalive.")
	flags.IntVarP(&o.Port, "port", "p", o.Port, "keepalive specifies the keep-alive period for an active network connection. Set to 0 to disable keepalive.")

	flags.StringVarP(&o.WWW, "www", "w", o.WWW, "Also serve static files from the given directory under the specified prefix.")
	flags.StringVarP(&o.WWWPrefix, "www-prefix", "P", o.WWWPrefix, "Prefix to serve static files under, if static file directory is specified.")
	flags.StringVar(&o.RejectMethods, "reject-methods", o.RejectMethods, "Regular expression for HTTP methods that the proxy should reject (example --reject-methods='POST,PUT,PATCH').")
	flags.StringVar(&o.RejectPaths, "reject-paths", o.RejectPaths, "Regular expression for paths that the proxy should reject. Paths specified here will be rejected even accepted by --accept-paths.")
	flags.StringVarP(&o.UNIXSocket, "unix-socket", "u", o.UNIXSocket, "Unix socket on which to run the proxy.")
}
