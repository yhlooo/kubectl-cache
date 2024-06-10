package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/sets"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewCacheProxyHandler 创建一个缓存代理 HTTP 处理器
func NewCacheProxyHandler(
	ctx context.Context,
	config *rest.Config,
	mapper meta.RESTMapper,
	apiProxyPrefix string,
) (*CacheProxyHandler, error) {
	logger := logr.FromContextOrDiscard(ctx)
	scheme := runtime.NewScheme()
	syncPeriod := 10 * time.Minute
	c, err := cache.New(config, cache.Options{
		Scheme:     scheme,
		Mapper:     mapper,
		SyncPeriod: &syncPeriod,
	})
	if err != nil {
		return nil, err
	}
	go func() {
		if err := c.Start(ctx); err != nil {
			logger.Error(err, "run cache error")
		}
	}()

	apisPathPrefix := strings.Trim(strings.Trim(apiProxyPrefix, "/")+"/apis", "/")
	legacyAPIsPathPrefix := strings.Trim(strings.Trim(apiProxyPrefix, "/")+"/api", "/")

	return &CacheProxyHandler{
		cache:  c,
		mapper: mapper,
		resolver: &apirequest.RequestInfoFactory{
			APIPrefixes:          sets.NewString(apisPathPrefix, legacyAPIsPathPrefix),
			GrouplessAPIPrefixes: sets.NewString(legacyAPIsPathPrefix),
		},
	}, nil
}

// CacheProxyHandler 缓存代理 HTTP 处理器
type CacheProxyHandler struct {
	cache    cache.Cache
	mapper   meta.RESTMapper
	resolver apirequest.RequestInfoResolver
}

var _ http.Handler = &CacheProxyHandler{}

// ServeHTTP 处理 HTTP 请求
func (h *CacheProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ret, err := h.Handle(req)
	if err != nil {
		logr.FromContextOrDiscard(req.Context()).Error(err, "handle request error")
		var apierr *apierrors.StatusError
		if errors.As(err, &apierr) {
			WriteResponse(w, int(apierr.Status().Code), apierr.Status())
			return
		}
		WriteResponse(w, http.StatusInternalServerError, apierrors.NewInternalError(err).Status())
		return
	}
	WriteResponse(w, http.StatusOK, ret)
}

// Handle 处理请求
func (h *CacheProxyHandler) Handle(req *http.Request) (runtime.Object, error) {
	// 检查请求
	info, err := h.resolver.NewRequestInfo(req)
	if err != nil {
		return nil, fmt.Errorf("resolve request error: %w", err)
	}
	gvr := schema.GroupVersionResource{
		Group:    info.APIGroup,
		Version:  info.APIVersion,
		Resource: info.Resource,
	}
	if info.Subresource != "" {
		if info.Subresource != "status" {
			return nil, apierrors.NewMethodNotSupported(gvr.GroupResource(), info.Verb)
		}
		gvr.Resource = info.Resource + "/" + info.Subresource
	}

	// 获取请求对应资源 Kind
	kind, err := h.mapper.KindFor(gvr)
	if err != nil {
		return nil, fmt.Errorf("get kind for %s error: %w", gvr.String(), err)
	}

	switch info.Verb {
	case "get":
		ret := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": kind.GroupVersion().String(),
				"kind":       kind.Kind,
			},
		}
		return ret, h.cache.Get(req.Context(), client.ObjectKey{
			Namespace: info.Namespace,
			Name:      info.Name,
		}, ret)
	case "list":
		ret := &unstructured.UnstructuredList{
			Object: map[string]interface{}{
				"apiVersion": kind.GroupVersion().String(),
				"kind":       kind.Kind + "List",
			},
		}
		var listOpts []client.ListOption
		if info.Namespace != "" {
			listOpts = append(listOpts, client.InNamespace(info.Namespace))
		}
		return ret, h.cache.List(req.Context(), ret, listOpts...)
	default:
		return nil, apierrors.NewMethodNotSupported(gvr.GroupResource(), info.Verb)
	}
}

// IsCached 判断该请求是否有缓存
func (h *CacheProxyHandler) IsCached(req *http.Request) bool {
	info, err := h.resolver.NewRequestInfo(req)
	if err != nil {
		// 解析请求错误
		return false
	}

	// 除 status 以外的子资源都不缓存
	if info.Subresource != "" && info.Subresource != "status" {
		return false
	}

	switch info.Verb {
	case "get", "list":
	default:
		// 其它请求都不缓存
		return false
	}

	// TODO: 检查部分参数

	return true
}

// WriteResponse 写响应
func WriteResponse(w http.ResponseWriter, code int, obj interface{}) {
	raw, err := json.Marshal(obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(raw)
}
