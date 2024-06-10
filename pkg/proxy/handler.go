package proxy

import (
	"net/http"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/proxy"
)

// NewProxyHandler 创建一个代理 HTTP 处理器
func NewProxyHandler(
	apiProxyPrefix string,
	filter *proxy.FilterServer,
	cfg *rest.Config,
	keepalive time.Duration,
	appendLocationPath bool,
) (http.Handler, error) {
	// 直连 handler
	passthrough, err := proxy.NewProxyHandler(apiProxyPrefix, filter, cfg, keepalive, appendLocationPath)
	if err != nil {
		return nil, err
	}

	return &proxyHandler{
		Handler: passthrough,
		cache:   NewCacheProxyHandler(cfg), // 缓存 handler
	}, nil
}

// proxyHandler 代理 HTTP 处理器
type proxyHandler struct {
	http.Handler

	cache *CacheProxyHandler
}

var _ http.Handler = &proxyHandler{}

// ServeHTTP 处理 HTTP 请求
func (h *proxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if h.cache == nil || !h.cache.IsCached(req) {
		// 直连
		h.Handler.ServeHTTP(w, req)
		return
	}

	// 缓存
	h.cache.ServeHTTP(w, req)
	return
}
