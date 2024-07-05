package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metascheme "k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/sets"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	registryrest "k8s.io/apiserver/pkg/registry/rest"
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
	AddKubernetesTypesToScheme(scheme)

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

	tableConvertor, err := NewDefaultTableConvertor(config, scheme)
	if err != nil {
		return nil, fmt.Errorf("create table convertor error: %w", err)
	}

	return &CacheProxyHandler{
		scheme: scheme,
		cache:  c,
		mapper: mapper,
		resolver: &apirequest.RequestInfoFactory{
			APIPrefixes:          sets.NewString(apisPathPrefix, legacyAPIsPathPrefix),
			GrouplessAPIPrefixes: sets.NewString(legacyAPIsPathPrefix),
		},
		tableConvertor: tableConvertor,
	}, nil
}

// CacheProxyHandler 缓存代理 HTTP 处理器
type CacheProxyHandler struct {
	scheme         *runtime.Scheme
	cache          cache.Cache
	mapper         meta.RESTMapper
	resolver       apirequest.RequestInfoResolver
	tableConvertor registryrest.TableConvertor

	startedInformersLock sync.RWMutex
	startedInformers     map[schema.GroupVersionResource]bool
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
	ctx := req.Context()
	logger := logr.FromContextOrDiscard(ctx)

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
	if info.Subresource != "" && info.Subresource != "status" {
		gvr.Resource = info.Resource + "/" + info.Subresource
		return nil, apierrors.NewMethodNotSupported(gvr.GroupResource(), info.Verb)
	}

	// 设置 informer
	if err := h.ensureInformer(ctx, gvr); err != nil {
		return nil, fmt.Errorf("ensure informer for %s error: %w", gvr, err)
	}

	// 获取请求对应资源 Kind
	gvk, err := h.mapper.KindFor(gvr)
	if err != nil {
		return nil, fmt.Errorf("get kind for %s error: %w", gvr.String(), err)
	}
	if info.Verb == "list" {
		gvk.Kind += "List"
	}

	// 创建返回对象
	obj, err := h.scheme.New(gvk)
	if err != nil {
		// 无结构对象
		obj = &unstructured.UnstructuredList{}
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)

	switch info.Verb {
	case "get":
		opts, err := ParseGetOptions(req)
		if err != nil {
			return nil, fmt.Errorf("parse get options error: %w", err)
		}
		ret, ok := obj.(client.Object)
		if !ok {
			return nil, fmt.Errorf("%T is not a client.Object", ret)
		}
		if err := h.HandleGet(ctx, ret, info.Namespace, info.Name, opts); err != nil {
			return nil, err
		}
	case "list":
		opts, err := ParseListOptions(req)
		if err != nil {
			return nil, fmt.Errorf("parse list options error: %w", err)
		}
		ret, ok := obj.(client.ObjectList)
		if !ok {
			return nil, fmt.Errorf("%T is not a client.ObjectList", ret)
		}
		if err := h.HandleList(ctx, ret, info.Namespace, opts); err != nil {
			return nil, err
		}
	default:
		return nil, apierrors.NewMethodNotSupported(gvr.GroupResource(), info.Verb)
	}

	// 转为列表
	accept := strings.Split(req.Header.Get("Accept"), ",")
	if !slices.Contains(accept, "application/json;as=Table;v=v1;g=meta.k8s.io") || h.tableConvertor == nil {
		// 不支持服务端表格，返回普通 json 格式
		return obj, nil
	}
	table, err := ConvertToTable(ctx, h.tableConvertor, obj)
	if err != nil {
		logger.V(1).Info(fmt.Sprintf("convert to table error: %v", err))
		return obj, nil
	}

	return table, nil
}

// IsCached 判断该请求是否有缓存
func (h *CacheProxyHandler) IsCached(req *http.Request) bool {
	info, err := h.resolver.NewRequestInfo(req)
	if err != nil {
		// 解析请求错误
		return false
	}

	// 没有具体资源的不缓存
	if info.Resource == "" {
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

// HandleGet 处理获取对象
func (h *CacheProxyHandler) HandleGet(
	ctx context.Context,
	obj client.Object,
	namespace, name string,
	opts metav1.GetOptions,
) error {
	return h.cache.Get(
		ctx,
		client.ObjectKey{Namespace: namespace, Name: name},
		obj,
		&client.GetOptions{Raw: &opts},
	)
}

// HandleList 处理列出对象
func (h *CacheProxyHandler) HandleList(
	ctx context.Context,
	obj client.ObjectList,
	namespace string,
	opts metav1.ListOptions,
) error {
	// 组装选项
	listOpts := &client.ListOptions{
		Namespace: namespace,
		//Limit:     opts.Limit, // TODO: 暂不支持分页，传递该选项会导致返回结果不完整
		//Continue:  opts.Continue,
		Raw: &opts,
	}
	if opts.LabelSelector != "" {
		selector, err := labels.Parse(opts.LabelSelector)
		if err != nil {
			return apierrors.NewBadRequest(err.Error())
		}
		listOpts.LabelSelector = selector
	}
	if opts.FieldSelector != "" {
		selector, err := fields.ParseSelector(opts.FieldSelector)
		if err != nil {
			return apierrors.NewBadRequest(err.Error())
		}
		listOpts.FieldSelector = selector
	}

	return h.cache.List(ctx, obj, listOpts)
}

// ensureInformer 确保资源对应 informer 就绪
func (h *CacheProxyHandler) ensureInformer(ctx context.Context, gvr schema.GroupVersionResource) error {
	logger := logr.FromContextOrDiscard(ctx)

	// 检查是否已经启动过对应的 informer
	h.startedInformersLock.RLock()
	if h.startedInformers[gvr] {
		h.startedInformersLock.RUnlock()
		return nil
	}

	// 换成写锁继续
	h.startedInformersLock.RUnlock()
	h.startedInformersLock.Lock()
	defer h.startedInformersLock.Unlock()

	// 换成写锁后再检查一遍，因为换锁过程中仍然有可能被修改
	if h.startedInformers[gvr] {
		return nil
	}

	gvk, err := h.mapper.KindFor(gvr)
	if err != nil {
		return fmt.Errorf("get kind for %s error: %w", gvr.String(), err)
	}
	obj, err := h.scheme.New(gvk)
	if err != nil {
		obj = &unstructured.Unstructured{}
		obj.GetObjectKind().SetGroupVersionKind(gvk)
	}

	if clientObj, ok := obj.(client.Object); ok {
		// 为对象设置字段索引
		if err := IndexFieldsForObject(ctx, h.cache, clientObj); err != nil {
			return fmt.Errorf("index fields for %T error: %w", obj, err)
		}
		// 创建 informer 并等待缓存同步
		logger.V(1).Info(fmt.Sprintf("waiting for informer for %s", gvk))
		if _, err := h.cache.GetInformer(ctx, clientObj, cache.BlockUntilSynced(true)); err != nil {
			return err
		}
	}

	if h.startedInformers == nil {
		h.startedInformers = make(map[schema.GroupVersionResource]bool)
	}
	h.startedInformers[gvr] = true

	return nil
}

// ConvertToTable 将 obj 转换为表格形式
func ConvertToTable(
	ctx context.Context,
	tableConvertor registryrest.TableConvertor,
	obj runtime.Object,
) (*metav1.Table, error) {
	// 转换为表格
	table, err := tableConvertor.ConvertToTable(ctx, obj, nil)
	if err != nil {
		return nil, err
	}
	table.GetObjectKind().SetGroupVersionKind(metav1.SchemeGroupVersion.WithKind("Table"))

	// 将每行 Object 转为 PartialObjectMetadata 或 PartialObjectMetadataList
	for i, row := range table.Rows {
		if row.Object.Object == nil || row.Object.Raw != nil {
			continue
		}
		partial, ok := ToPartial(row.Object.Object)
		if !ok {
			continue
		}
		table.Rows[i].Object.Object = partial
	}

	return table, nil
}

// WriteResponse 写响应
func WriteResponse(w http.ResponseWriter, code int, obj interface{}) {
	if status, ok := obj.(metav1.Status); ok {
		status.APIVersion = "v1"
		status.Kind = "Status"
		obj = status
	}

	raw, err := json.Marshal(obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(raw)
}

// ParseGetOptions 解析请求 Get 选项
func ParseGetOptions(req *http.Request) (metav1.GetOptions, error) {
	ret := metav1.GetOptions{}
	values := req.URL.Query()
	if len(values) > 0 {
		if err := metascheme.ParameterCodec.DecodeParameters(values, metav1.SchemeGroupVersion, &ret); err != nil {
			return ret, err
		}
	}
	return ret, nil
}

// ParseListOptions 解析请求 List 选项
func ParseListOptions(req *http.Request) (metav1.ListOptions, error) {
	ret := metav1.ListOptions{}
	values := req.URL.Query()
	if len(values) > 0 {
		if err := metascheme.ParameterCodec.DecodeParameters(values, metav1.SchemeGroupVersion, &ret); err != nil {
			return ret, err
		}
	}
	return ret, nil
}

// ToPartial 从对象提取仅包含 metadata 的部分
func ToPartial(obj runtime.Object) (runtime.Object, bool) {
	switch typedObj := obj.(type) {
	case metav1.ObjectMetaAccessor:
		//goland:noinspection GoDeprecation
		return &metav1.PartialObjectMetadata{
			TypeMeta: metav1.TypeMeta{
				APIVersion: metav1.SchemeGroupVersion.String(),
				Kind:       "PartialObjectMetadata",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:                       typedObj.GetObjectMeta().GetName(),
				GenerateName:               typedObj.GetObjectMeta().GetGenerateName(),
				Namespace:                  typedObj.GetObjectMeta().GetNamespace(),
				SelfLink:                   typedObj.GetObjectMeta().GetSelfLink(),
				UID:                        typedObj.GetObjectMeta().GetUID(),
				ResourceVersion:            typedObj.GetObjectMeta().GetResourceVersion(),
				Generation:                 typedObj.GetObjectMeta().GetGeneration(),
				CreationTimestamp:          typedObj.GetObjectMeta().GetCreationTimestamp(),
				DeletionTimestamp:          typedObj.GetObjectMeta().GetDeletionTimestamp(),
				DeletionGracePeriodSeconds: typedObj.GetObjectMeta().GetDeletionGracePeriodSeconds(),
				Labels:                     typedObj.GetObjectMeta().GetLabels(),
				Annotations:                typedObj.GetObjectMeta().GetAnnotations(),
				OwnerReferences:            typedObj.GetObjectMeta().GetOwnerReferences(),
				Finalizers:                 typedObj.GetObjectMeta().GetFinalizers(),
				ManagedFields:              typedObj.GetObjectMeta().GetManagedFields(),
			},
		}, true
	case metav1.ListMetaAccessor:
		//goland:noinspection GoDeprecation
		return &metav1.PartialObjectMetadataList{
			TypeMeta: metav1.TypeMeta{
				APIVersion: metav1.SchemeGroupVersion.String(),
				Kind:       "PartialObjectMetadataList",
			},
			ListMeta: metav1.ListMeta{
				SelfLink:           typedObj.GetListMeta().GetSelfLink(),
				ResourceVersion:    typedObj.GetListMeta().GetResourceVersion(),
				Continue:           typedObj.GetListMeta().GetContinue(),
				RemainingItemCount: typedObj.GetListMeta().GetRemainingItemCount(),
			},
		}, true
	}
	return nil, false
}
