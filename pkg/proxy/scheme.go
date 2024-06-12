package proxy

import (
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/runtime"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/pkg/apis/core"
)

// AddKubernetesTypesToScheme 添加 Kubernetes 资源到 scheme
func AddKubernetesTypesToScheme(scheme *runtime.Scheme) {
	_ = clientscheme.AddToScheme(scheme)
	_ = metainternalversion.AddToScheme(scheme)
	_ = core.AddToScheme(scheme)
}
