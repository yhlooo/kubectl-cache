package options

// NewDefaultOptions 创建一个默认运行选项
func NewDefaultOptions() Options {
	return Options{
		Global: NewDefaultGlobalOptions(),
		Proxy:  NewDefaultProxyOptions(),
	}
}

// Options pcrctl 运行选项
type Options struct {
	// 全局选项
	Global GlobalOptions
	// proxy 命令选项
	Proxy ProxyOptions
}
