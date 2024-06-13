package options

import (
	"github.com/spf13/pflag"
)

// NewDefaultInternalProxyOptions 创建一个默认的 internal-proxy 子命令选项
func NewDefaultInternalProxyOptions() InternalProxyOptions {
	return InternalProxyOptions{}
}

// InternalProxyOptions internal-proxy 子命令选项
type InternalProxyOptions struct{}

// AddPFlags 将选项绑定到命令行参数
func (opts *InternalProxyOptions) AddPFlags(_ *pflag.FlagSet) {}
