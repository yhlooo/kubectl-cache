package options

import "github.com/spf13/pflag"

// NewDefaultShutdownOptions 创建一个默认的 shutdown 子命令选项
func NewDefaultShutdownOptions() ShutdownOptions {
	return ShutdownOptions{
		Wait:  true,
		Force: false,
		All:   false,
	}
}

// ShutdownOptions shutdown 子命令选项
type ShutdownOptions struct {
	// 等待代理退出
	Wait bool
	// 强制退出
	Force bool
	// 退出所有代理
	All bool
}

// AddPFlags 将选项绑定到命令行
func (opts *ShutdownOptions) AddPFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&opts.Wait, "wait", opts.Wait, "Wait for proxy to be shutdown.")
	flags.BoolVar(&opts.Force, "force", opts.Force, "Force shutdown.")
	flags.BoolVarP(&opts.All, "all", "A", opts.All, "Shutdown all proxies.")
}
