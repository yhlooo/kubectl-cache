package options

// NewDefaultOptions 创建一个默认运行选项
func NewDefaultOptions() Options {
	return Options{
		Global:               NewDefaultGlobalOptions(),
		Proxy:                NewDefaultProxyOptions(),
		Proxies:              NewDefaultProxiesOptions(),
		InternalProxyOptions: NewDefaultInternalProxyOptions(),
	}
}

// Options pcrctl 运行选项
type Options struct {
	// 全局选项
	Global GlobalOptions
	// proxy 子命令选项
	Proxy ProxyOptions
	// proxies 子命令选项
	Proxies ProxiesOptions

	// internal-proxy 子命令选项
	InternalProxyOptions InternalProxyOptions
}
