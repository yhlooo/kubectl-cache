package proxy

import (
	"net/http"

	"k8s.io/client-go/rest"
)

// NewCacheProxyHandler 创建一个缓存代理 HTTP 处理器
func NewCacheProxyHandler(config *rest.Config) *CacheProxyHandler {
	//TODO implement me
	return nil
}

// CacheProxyHandler 缓存代理 HTTP 处理器
type CacheProxyHandler struct{}

var _ http.Handler = &CacheProxyHandler{}

// ServeHTTP 处理 HTTP 请求
func (h *CacheProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//TODO implement me
	panic("implement me")
}

// IsCached 判断该请求是否有缓存
func (h *CacheProxyHandler) IsCached(req *http.Request) bool {
	// TODO: ...
	return false
}
