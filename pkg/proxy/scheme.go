package proxy

import (
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/runtime"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/pkg/apis/abac"
	"k8s.io/kubernetes/pkg/apis/admission"
	"k8s.io/kubernetes/pkg/apis/admissionregistration"
	"k8s.io/kubernetes/pkg/apis/apidiscovery"
	"k8s.io/kubernetes/pkg/apis/apiserverinternal"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/apis/authentication"
	"k8s.io/kubernetes/pkg/apis/authorization"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/certificates"
	"k8s.io/kubernetes/pkg/apis/coordination"
	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/discovery"
	"k8s.io/kubernetes/pkg/apis/events"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/apis/flowcontrol"
	"k8s.io/kubernetes/pkg/apis/imagepolicy"
	"k8s.io/kubernetes/pkg/apis/networking"
	"k8s.io/kubernetes/pkg/apis/node"
	"k8s.io/kubernetes/pkg/apis/policy"
	"k8s.io/kubernetes/pkg/apis/rbac"
	"k8s.io/kubernetes/pkg/apis/resource"
	"k8s.io/kubernetes/pkg/apis/scheduling"
	"k8s.io/kubernetes/pkg/apis/storage"
	"k8s.io/kubernetes/pkg/apis/storagemigration"
)

// AddKubernetesTypesToScheme 添加 Kubernetes 资源到 scheme
func AddKubernetesTypesToScheme(scheme *runtime.Scheme) {
	_ = clientscheme.AddToScheme(scheme)
	_ = metainternalversion.AddToScheme(scheme)

	_ = abac.AddToScheme(scheme)
	_ = admission.AddToScheme(scheme)
	_ = admissionregistration.AddToScheme(scheme)
	_ = apidiscovery.AddToScheme(scheme)
	_ = apiserverinternal.AddToScheme(scheme)
	_ = apps.AddToScheme(scheme)
	_ = authentication.AddToScheme(scheme)
	_ = authorization.AddToScheme(scheme)
	_ = autoscaling.AddToScheme(scheme)
	_ = batch.AddToScheme(scheme)
	_ = certificates.AddToScheme(scheme)
	_ = coordination.AddToScheme(scheme)
	_ = core.AddToScheme(scheme)
	_ = discovery.AddToScheme(scheme)
	_ = events.AddToScheme(scheme)
	_ = extensions.AddToScheme(scheme)
	_ = flowcontrol.AddToScheme(scheme)
	_ = imagepolicy.AddToScheme(scheme)
	_ = networking.AddToScheme(scheme)
	_ = node.AddToScheme(scheme)
	_ = policy.AddToScheme(scheme)
	_ = rbac.AddToScheme(scheme)
	_ = resource.AddToScheme(scheme)
	_ = scheduling.AddToScheme(scheme)
	_ = storage.AddToScheme(scheme)
	_ = storagemigration.AddToScheme(scheme)
}
