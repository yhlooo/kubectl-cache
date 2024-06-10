package options

import (
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewDefaultGlobalOptions 返回默认全局选项
func NewDefaultGlobalOptions() GlobalOptions {
	return GlobalOptions{
		Verbosity:    0,
		ClientConfig: genericclioptions.NewConfigFlags(true),
	}
}

// GlobalOptions 全局选项
type GlobalOptions struct {
	// Kubernetes 客户端配置选项
	ClientConfig *genericclioptions.ConfigFlags
	// 日志数量级别（ 0 / 1 / 2 ）
	Verbosity uint32
}

// Validate 校验选项是否合法
func (o *GlobalOptions) Validate() error {
	if o.Verbosity > 2 {
		return fmt.Errorf("invalid log verbosity: %d (expected: 0, 1 or 2)", o.Verbosity)
	}
	return nil
}

// AddPFlags 将选项绑定到命令行参数
func (o *GlobalOptions) AddPFlags(flags *pflag.FlagSet) {
	o.ClientConfig.AddFlags(flags)
	flags.Uint32VarP(&o.Verbosity, "v", "v", o.Verbosity, "Number for the log level verbosity (0, 1, or 2)")
}
