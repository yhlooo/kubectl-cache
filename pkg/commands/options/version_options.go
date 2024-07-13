package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// NewDefaultVersionOptions 创建一个默认的 version 子命令选项
func NewDefaultVersionOptions() VersionOptions {
	return VersionOptions{
		OutputFormat: "",
	}
}

// VersionOptions version 子命令选项
type VersionOptions struct {
	// 输出格式
	// yaml 或 json
	OutputFormat string
}

// Validate 校验选项
func (opts *VersionOptions) Validate() error {
	switch opts.OutputFormat {
	case "", "yaml", "json":
	default:
		return fmt.Errorf("invalid output format: %s (must be one of 'yaml' or 'json')", opts.OutputFormat)
	}
	return nil
}

// AddPFlags 将选项绑定到命令行
func (opts *VersionOptions) AddPFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&opts.OutputFormat, "output", "o", opts.OutputFormat, "Output format. One of 'yaml' or 'json'.")
}
