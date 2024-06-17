package options

import (
	"time"

	"github.com/spf13/pflag"
)

// NewDefaultInternalProxyOptions 创建一个默认的 internal-proxy 子命令选项
func NewDefaultInternalProxyOptions() InternalProxyOptions {
	return InternalProxyOptions{
		MaxIdleTime: 10 * time.Minute,
	}
}

// InternalProxyOptions internal-proxy 子命令选项
type InternalProxyOptions struct {
	// 最大空闲时间（超过后代理服务自行关闭）
	MaxIdleTime time.Duration
}

// AddPFlags 将选项绑定到命令行参数
func (opts *InternalProxyOptions) AddPFlags(flags *pflag.FlagSet) {
	flags.DurationVar(&opts.MaxIdleTime, "max-idle-time", opts.MaxIdleTime, "Max idle time. (0 means no limit)")
}
