package proxy

import (
	"context"
	"fmt"
	"sync"
	"time"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	informersexternalversions "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	listersapiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1"
	customresourcetableconvertor "k8s.io/apiextensions-apiserver/pkg/registry/customresource/tableconvertor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	registryrest "k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/apis/admissionregistration"
	"k8s.io/kubernetes/pkg/apis/apiserverinternal"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/certificates"
	"k8s.io/kubernetes/pkg/apis/coordination"
	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/discovery"
	"k8s.io/kubernetes/pkg/apis/flowcontrol"
	"k8s.io/kubernetes/pkg/apis/networking"
	"k8s.io/kubernetes/pkg/apis/node"
	"k8s.io/kubernetes/pkg/apis/policy"
	"k8s.io/kubernetes/pkg/apis/rbac"
	"k8s.io/kubernetes/pkg/apis/resource"
	"k8s.io/kubernetes/pkg/apis/scheduling"
	"k8s.io/kubernetes/pkg/apis/storage"
	"k8s.io/kubernetes/pkg/apis/storagemigration"
	"k8s.io/kubernetes/pkg/printers"
	printersinternal "k8s.io/kubernetes/pkg/printers/internalversion"
	printerstorage "k8s.io/kubernetes/pkg/printers/storage"
)

// TableConvertorGetter TableConvertor 获取器
type TableConvertorGetter interface {
	// TableConvertorForObject 获取指定对象的表格转换器
	TableConvertorForObject(obj runtime.Object) (registryrest.TableConvertor, error)
}

// TableConvertorGetterFunc 获取指定对象的表格转换器方法
type TableConvertorGetterFunc func(obj runtime.Object) (registryrest.TableConvertor, error)

var _ TableConvertorGetter = TableConvertorGetterFunc(nil)

// TableConvertorForObject 获取指定对象的表格转换器
func (getterFunc TableConvertorGetterFunc) TableConvertorForObject(
	obj runtime.Object,
) (registryrest.TableConvertor, error) {
	return getterFunc(obj)
}

// TableConvertorGetters 是 TableConvertorGetter 的列表
type TableConvertorGetters []TableConvertorGetter

var _ TableConvertorGetter = TableConvertorGetters(nil)

// TableConvertorForObject 获取指定对象的表格转换器
func (getters TableConvertorGetters) TableConvertorForObject(
	obj runtime.Object,
) (registryrest.TableConvertor, error) {
	for _, getter := range getters {
		if convertor, err := getter.TableConvertorForObject(obj); err == nil {
			return convertor, err
		}
	}
	return nil, fmt.Errorf("no table convertor found for %s", obj.GetObjectKind().GroupVersionKind())
}

// NewCachedTableConvertorGetter 创建一个带缓存的 TableConvertorGetter
func NewCachedTableConvertorGetter(getter TableConvertorGetter) TableConvertorGetter {
	return &cachedTableConvertorGetter{
		TableConvertorGetter: getter,
	}
}

// cachedTableConvertorGetter 带缓存的 TableConvertorGetter
type cachedTableConvertorGetter struct {
	TableConvertorGetter

	lock  sync.RWMutex
	cache map[schema.GroupVersionKind]registryrest.TableConvertor
}

var _ TableConvertorGetter = &cachedTableConvertorGetter{}

// TableConvertorForObject 获取指定对象的表格转换器
func (getter *cachedTableConvertorGetter) TableConvertorForObject(
	obj runtime.Object,
) (registryrest.TableConvertor, error) {
	gvk := obj.GetObjectKind().GroupVersionKind()

	// 尝试从缓存获取
	getter.lock.RLock()
	if cachedGetter, ok := getter.cache[gvk]; ok && cachedGetter != nil {
		getter.lock.RUnlock()
		return cachedGetter, nil
	}
	getter.lock.RUnlock()

	ret, err := getter.TableConvertorGetter.TableConvertorForObject(obj)
	if err != nil {
		return ret, err
	}

	// 更新缓存
	getter.lock.Lock()
	defer getter.lock.Unlock()
	if getter.cache == nil {
		getter.cache = make(map[schema.GroupVersionKind]registryrest.TableConvertor)
	}
	getter.cache[gvk] = ret

	return ret, nil
}

// NewCRDTableConvertorGetterForConfig 基于客户端配置创建一个 CRD TableConvertorGetter
func NewCRDTableConvertorGetterForConfig(config *rest.Config) (TableConvertorGetter, error) {
	client, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	factory := informersexternalversions.NewSharedInformerFactoryWithOptions(client, time.Hour)
	lister := factory.Apiextensions().V1().CustomResourceDefinitions().Lister()
	factory.Start(nil)
	return NewCRDTableConvertorGetter(lister), nil
}

// NewCRDTableConvertorGetter 创建一个 CRD TableConvertorGetter
func NewCRDTableConvertorGetter(
	crdGetter listersapiextensionsv1.CustomResourceDefinitionLister,
) TableConvertorGetter {
	return &crdTableConvertorGetter{
		crdGetter: crdGetter,
	}
}

// crdTableConvertorGetter CRD TableConvertorGetter
type crdTableConvertorGetter struct {
	crdGetter listersapiextensionsv1.CustomResourceDefinitionLister
}

var _ TableConvertorGetter = &crdTableConvertorGetter{}

// TableConvertorForObject 获取指定对象的 TableConvertor
func (getter *crdTableConvertorGetter) TableConvertorForObject(
	obj runtime.Object,
) (registryrest.TableConvertor, error) {
	crds, err := getter.crdGetter.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	gvk := obj.GetObjectKind().GroupVersionKind()
	for _, crd := range crds {
		if crd.Spec.Group != gvk.Group {
			continue
		}
		if crd.Spec.Names.Kind != gvk.Kind && crd.Spec.Names.ListKind != gvk.Kind {
			continue
		}
		for _, version := range crd.Spec.Versions {
			if version.Name != gvk.Version {
				continue
			}
			return customresourcetableconvertor.New(version.AdditionalPrinterColumns)
		}
	}
	return nil, fmt.Errorf("no crd found for %s", gvk.GroupKind().String())
}

// AggregateTableConvertor 聚合的 TableConvertor
type AggregateTableConvertor struct {
	TableConvertorGetter TableConvertorGetter
}

var _ registryrest.TableConvertor = AggregateTableConvertor{}

// ConvertToTable 转换为表格
func (tc AggregateTableConvertor) ConvertToTable(
	ctx context.Context,
	obj runtime.Object,
	opts runtime.Object,
) (*metav1.Table, error) {
	convertor, err := tc.TableConvertorGetter.TableConvertorForObject(obj)
	if err != nil {
		return nil, fmt.Errorf("get convertor for %T error: %w", obj, err)
	}
	return convertor.ConvertToTable(ctx, obj, opts)
}

// MultiVersionTableConvertor 基于 __internal 版本资源的 TableConvertor 的可转换多版本资源的 TableConvertor
type MultiVersionTableConvertor struct {
	IntervalVersion registryrest.TableConvertor
	Scheme          *runtime.Scheme
}

var _ registryrest.TableConvertor = MultiVersionTableConvertor{}

// ConvertToTable 转换为表格
func (tc MultiVersionTableConvertor) ConvertToTable(
	ctx context.Context,
	obj runtime.Object,
	opts runtime.Object,
) (*metav1.Table, error) {
	// 转换为 __internal 版本
	gvk := obj.GetObjectKind().GroupVersionKind()
	internalObj, err := tc.Scheme.New(gvk.GroupKind().WithVersion(runtime.APIVersionInternal))
	if err != nil {
		return nil, fmt.Errorf("new object with internal version for %T error: %w", obj, err)
	}
	if err := tc.Scheme.Convert(obj, internalObj, nil); err != nil {
		return nil, fmt.Errorf("convert %T to internal version error: %w", obj, err)
	}

	// 转换为表格
	return tc.IntervalVersion.ConvertToTable(ctx, internalObj, opts)
}

// BuiltinTableConvertorGetter 内置对象 TableConvertor 的 TableConvertorGetter
func BuiltinTableConvertorGetter(scheme *runtime.Scheme) TableConvertorGetter {
	builtinConvertor := MultiVersionTableConvertor{
		IntervalVersion: printerstorage.TableConvertor{
			TableGenerator: printers.NewTableGenerator().With(printersinternal.AddHandlers),
		},
		Scheme: scheme,
	}
	return TableConvertorGetterFunc(func(obj runtime.Object) (registryrest.TableConvertor, error) {
		switch obj.GetObjectKind().GroupVersionKind().Group {
		case core.GroupName,
			apps.GroupName,
			policy.GroupName,
			batch.GroupName,
			networking.GroupName,
			autoscaling.GroupName,
			rbac.GroupName,
			certificates.GroupName,
			coordination.GroupName,
			storage.GroupName,
			metav1.GroupName,
			scheduling.GroupName,
			node.GroupName,
			discovery.GroupName,
			admissionregistration.GroupName,
			flowcontrol.GroupName,
			apiserverinternal.GroupName,
			resource.GroupName,
			storagemigration.GroupName:
			return builtinConvertor, nil
		}
		return nil, fmt.Errorf("builtin table convertor not support %s", obj.GetObjectKind().GroupVersionKind())
	})
}

// NewDefaultTableConvertor 创建一个默认的 TableConvertor
func NewDefaultTableConvertor(config *rest.Config, scheme *runtime.Scheme) (registryrest.TableConvertor, error) {
	crdConvertorGetter, err := NewCRDTableConvertorGetterForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create crd table convertor getter error: %w", err)
	}

	return AggregateTableConvertor{
		TableConvertorGetter: TableConvertorGetters{
			BuiltinTableConvertorGetter(scheme),
			NewCachedTableConvertorGetter(crdConvertorGetter),
		},
	}, nil
}
