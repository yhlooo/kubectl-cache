package proxy

import (
	apiextensionsinstall "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/runtime"
	apiregistrationinstall "k8s.io/kube-aggregator/pkg/apis/apiregistration/install"
	admissioninstall "k8s.io/kubernetes/pkg/apis/admission/install"
	admissionregistrationinstall "k8s.io/kubernetes/pkg/apis/admissionregistration/install"
	apiserverinternalinstall "k8s.io/kubernetes/pkg/apis/apiserverinternal/install"
	appsinstall "k8s.io/kubernetes/pkg/apis/apps/install"
	authenticationinstall "k8s.io/kubernetes/pkg/apis/authentication/install"
	authorizationinstall "k8s.io/kubernetes/pkg/apis/authorization/install"
	autoscalinginstall "k8s.io/kubernetes/pkg/apis/autoscaling/install"
	batchinstall "k8s.io/kubernetes/pkg/apis/batch/install"
	certificatesinstall "k8s.io/kubernetes/pkg/apis/certificates/install"
	coordinationinstall "k8s.io/kubernetes/pkg/apis/coordination/install"
	coreinstall "k8s.io/kubernetes/pkg/apis/core/install"
	discoveryinstall "k8s.io/kubernetes/pkg/apis/discovery/install"
	eventsinstall "k8s.io/kubernetes/pkg/apis/events/install"
	extensionsinstall "k8s.io/kubernetes/pkg/apis/extensions/install"
	flowcontrolinstall "k8s.io/kubernetes/pkg/apis/flowcontrol/install"
	imagepolicyinstall "k8s.io/kubernetes/pkg/apis/imagepolicy/install"
	networkinginstall "k8s.io/kubernetes/pkg/apis/networking/install"
	nodeinstall "k8s.io/kubernetes/pkg/apis/node/install"
	policyinstall "k8s.io/kubernetes/pkg/apis/policy/install"
	rbacinstall "k8s.io/kubernetes/pkg/apis/rbac/install"
	resourceinstall "k8s.io/kubernetes/pkg/apis/resource/install"
	schedulinginstall "k8s.io/kubernetes/pkg/apis/scheduling/install"
	storageinstall "k8s.io/kubernetes/pkg/apis/storage/install"
	storagemigrationinstall "k8s.io/kubernetes/pkg/apis/storagemigration/install"
)

// AddKubernetesTypesToScheme 添加 Kubernetes 资源到 scheme
func AddKubernetesTypesToScheme(scheme *runtime.Scheme) {
	_ = metainternalversion.AddToScheme(scheme)

	admissioninstall.Install(scheme)
	admissionregistrationinstall.Install(scheme)
	apiserverinternalinstall.Install(scheme)
	appsinstall.Install(scheme)
	authenticationinstall.Install(scheme)
	authorizationinstall.Install(scheme)
	autoscalinginstall.Install(scheme)
	batchinstall.Install(scheme)
	certificatesinstall.Install(scheme)
	coordinationinstall.Install(scheme)
	coreinstall.Install(scheme)
	discoveryinstall.Install(scheme)
	eventsinstall.Install(scheme)
	extensionsinstall.Install(scheme)
	flowcontrolinstall.Install(scheme)
	imagepolicyinstall.Install(scheme)
	networkinginstall.Install(scheme)
	nodeinstall.Install(scheme)
	policyinstall.Install(scheme)
	rbacinstall.Install(scheme)
	resourceinstall.Install(scheme)
	schedulinginstall.Install(scheme)
	storageinstall.Install(scheme)
	storagemigrationinstall.Install(scheme)

	apiextensionsinstall.Install(scheme)
	apiregistrationinstall.Install(scheme)
}
