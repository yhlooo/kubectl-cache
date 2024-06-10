package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/proxy"
)

// NewProxyHandler 创建一个代理 HTTP 处理器
func NewProxyHandler(
	ctx context.Context,
	apiProxyPrefix string,
	filter *proxy.FilterServer,
	cfg *rest.Config,
	mapper meta.RESTMapper,
	keepalive time.Duration,
	appendLocationPath bool,
) (http.Handler, error) {
	logger := logr.FromContextOrDiscard(ctx)

	// 直连 handler
	passthrough, err := proxy.NewProxyHandler(apiProxyPrefix, nil, cfg, keepalive, appendLocationPath)
	if err != nil {
		return nil, err
	}

	// 缓存 handler
	cache, err := NewCacheProxyHandler(ctx, cfg, mapper, apiProxyPrefix)
	if err != nil {
		return nil, err
	}

	h := http.Handler(&proxyHandler{
		Handler: passthrough,
		logger:  logger,
		cache:   cache, // 缓存 handler
	})

	// 添加过滤器
	if filter != nil {
		h = filter.HandlerFor(h)
	}

	return h, nil
}

// proxyHandler 代理 HTTP 处理器
type proxyHandler struct {
	http.Handler

	logger logr.Logger

	cacheFilterHandler http.Handler
	cache              *CacheProxyHandler
}

var _ http.Handler = &proxyHandler{}

// ServeHTTP 处理 HTTP 请求
func (h *proxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger := h.logger
	req = req.WithContext(logr.NewContext(req.Context(), logger))

	if h.cache == nil || !h.cache.IsCached(req) {
		// 直连
		logger.V(1).Info(fmt.Sprintf("PASSTHROUGH %s %s", req.Method, req.RequestURI))
		h.Handler.ServeHTTP(w, req)
		return
	}

	// 缓存
	logger.V(1).Info(fmt.Sprintf("CACHED      %s %s", req.Method, req.RequestURI))
	h.cache.ServeHTTP(w, req)
	return
}
