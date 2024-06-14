package proxyclientgetter

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"

	"github.com/yhlooo/kubectl-cache/pkg/proxymgr"
)

// ProxyClientGetter 代理客户端获取器
type ProxyClientGetter struct {
	genericclioptions.RESTClientGetter

	ProxyManager proxymgr.ProxyManager

	ctx context.Context
}

var _ genericclioptions.RESTClientGetter = &ProxyClientGetter{}

// SetContext 设置上下文
func (getter *ProxyClientGetter) SetContext(ctx context.Context) {
	getter.ctx = ctx
}

// ToRESTConfig 获取客户端配置
func (getter *ProxyClientGetter) ToRESTConfig() (*rest.Config, error) {
	ctx := getter.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	logger := logr.FromContextOrDiscard(ctx)

	// 获取原始配置
	config, err := getter.RESTClientGetter.ToRESTConfig()
	if err != nil || getter.ProxyManager == nil {
		return config, err
	}

	// 尝试获取正在运行的代理
	info, err := getter.ProxyManager.GetForConfig(ctx, config)
	if err != nil {
		// 没有的话新启动一个
		info, err = getter.ProxyManager.NewForConfig(ctx, config)
		if err != nil {
			logger.Info(fmt.Sprintf("WARNING start cache proxy error, use passthrough mode, error: %v", err))
			return config, nil
		}
	}

	logger.Info(fmt.Sprintf("using proxy http://127.0.0.1:%d", info.TCPPort()))
	return info.ToClientConfig(), nil
}
