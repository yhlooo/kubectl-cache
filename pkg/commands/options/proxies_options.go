package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// NewDefaultProxiesOptions 创建一个默认的 proxies 子命令选项
func NewDefaultProxiesOptions() ProxiesOptions {
	return ProxiesOptions{
		OutputFormat: "",
	}
}

// ProxiesOptions proxies 子命令选项
type ProxiesOptions struct {
	// 输出格式
	OutputFormat string
}

// Validate 校验选项
func (opts *ProxiesOptions) Validate() error {
	switch opts.OutputFormat {
	case "", "json", "yaml":
	default:
		return fmt.Errorf("unsupported output format %q, must be one of \"json\" or \"yaml\" )", opts.OutputFormat)
	}
	return nil
}

// AddPFlags 将选项绑定到命令行
func (opts *ProxiesOptions) AddPFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&opts.OutputFormat, "output", "o", opts.OutputFormat, "Output format. One of: (json, yaml)")
}
