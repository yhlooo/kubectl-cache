package proxy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IndexFieldsForObject 为对象设置字段索引
func IndexFieldsForObject(ctx context.Context, c cache.Cache, obj client.Object) error {
	// 添加 metadata 字段索引
	if err := IndexFieldsForObjectMeta(ctx, c, obj); err != nil {
		return err
	}

	// 添加额外字段索引
	switch obj.(type) {
	case *appsv1.ReplicaSet:
		return IndexFieldsForAppsV1ReplicaSet(ctx, c)
	case *corev1.Event:
		return IndexFieldsForCoreV1Event(ctx, c)
	case *corev1.Pod:
		return IndexFieldsForCoreV1Pod(ctx, c)
	}

	return nil
}

// IndexFieldsForObjectMeta 为任意资源添加 metadata 字段索引
func IndexFieldsForObjectMeta(ctx context.Context, c cache.Cache, obj client.Object) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.V(1).Info(fmt.Sprintf("set index field \"metadata.name\" for %T", obj))
	if err := c.IndexField(ctx, obj, "metadata.name", getObjectNames); err != nil {
		return err
	}
	logger.V(1).Info(fmt.Sprintf("set index field \"metadata.namespace\" for %T", obj))
	if err := c.IndexField(ctx, obj, "metadata.namespace", getObjectNamespaces); err != nil {
		return err
	}
	return nil
}

// getObjectNames 获取对象名
func getObjectNames(obj client.Object) []string {
	return []string{obj.GetName()}
}

// getObjectNamespaces 获取对象命名空间
func getObjectNamespaces(obj client.Object) []string {
	return []string{obj.GetNamespace()}
}

// IndexFieldsForCoreV1Pod 为 corev1.Pod 设置字段索引
func IndexFieldsForCoreV1Pod(ctx context.Context, c cache.Cache) error {
	logger := logr.FromContextOrDiscard(ctx)
	for _, field := range []string{
		"spec.nodeName",
		"spec.restartPolicy",
		"spec.schedulerName",
		"spec.serviceAccountName",
		"spec.hostNetwork",
		"status.podIP",
		"status.phase",
		"status.nominatedNodeName",
	} {
		logger.V(1).Info(fmt.Sprintf("set index field %q for *v1.Pod", field))
		if err := c.IndexField(ctx, &corev1.Pod{}, field, getCoreV1PodField(field)); err != nil {
			return err
		}
	}
	return nil
}

// getCoreV1PodField 获取 corev1.Pod 字段方法
func getCoreV1PodField(field string) client.IndexerFunc {
	return func(obj client.Object) []string {
		pod, ok := obj.(*corev1.Pod)
		if !ok || pod == nil {
			return nil
		}
		val := ""
		switch field {
		case "spec.nodeName":
			val = pod.Spec.NodeName
		case "spec.restartPolicy":
			val = string(pod.Spec.RestartPolicy)
		case "spec.schedulerName":
			val = pod.Spec.SchedulerName
		case "spec.serviceAccountName":
			val = pod.Spec.ServiceAccountName
		case "spec.hostNetwork":
			val = strconv.FormatBool(pod.Spec.HostNetwork)
		case "status.podIP":
			val = pod.Status.PodIP
		case "status.phase":
			val = string(pod.Status.Phase)
		case "status.nominatedNodeName":
			val = pod.Status.NominatedNodeName
		default:
			return nil
		}
		return []string{val}
	}
}

// IndexFieldsForCoreV1Event 为 corev1.Event 设置字段索引
func IndexFieldsForCoreV1Event(ctx context.Context, c cache.Cache) error {
	logger := logr.FromContextOrDiscard(ctx)
	for _, field := range []string{
		"involvedObject.kind",
		"involvedObject.namespace",
		"involvedObject.name",
		"involvedObject.uid",
		"involvedObject.apiVersion",
		"involvedObject.resourceVersion",
		"involvedObject.fieldPath",
		"reason",
		"reportingComponent",
		"source",
		"type",
	} {
		logger.V(1).Info(fmt.Sprintf("set index field %q for *v1.Event", field))
		if err := c.IndexField(ctx, &corev1.Event{}, field, getCoreV1EventField(field)); err != nil {
			return err
		}
	}
	return nil
}

// getCoreV1EventField 获取 corev1.Event 字段方法
func getCoreV1EventField(field string) client.IndexerFunc {
	return func(obj client.Object) []string {
		event, ok := obj.(*corev1.Event)
		if !ok || event == nil {
			return nil
		}
		val := ""
		switch field {
		case "involvedObject.kind":
			val = event.InvolvedObject.Kind
		case "involvedObject.namespace":
			val = event.InvolvedObject.Namespace
		case "involvedObject.name":
			val = event.InvolvedObject.Name
		case "involvedObject.uid":
			val = string(event.InvolvedObject.UID)
		case "involvedObject.apiVersion":
			val = event.InvolvedObject.APIVersion
		case "involvedObject.resourceVersion":
			val = event.InvolvedObject.ResourceVersion
		case "involvedObject.fieldPath":
			val = event.InvolvedObject.FieldPath
		case "reason":
			val = event.Reason
		case "reportingComponent":
			val = event.ReportingController
		case "source":
			val = event.Source.Component
			if val == "" {
				val = event.ReportingController
			}
		case "type":
			val = event.Type
		default:
			return nil
		}
		return []string{val}
	}
}

// IndexFieldsForAppsV1ReplicaSet 为 appsv1.ReplicaSet 设置字段索引
func IndexFieldsForAppsV1ReplicaSet(ctx context.Context, c cache.Cache) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.V(1).Info("set index field \"status.replicas\" for *v1.ReplicaSet")
	if err := c.IndexField(ctx, &appsv1.ReplicaSet{}, "status.replicas", func(obj client.Object) []string {
		rs, ok := obj.(*appsv1.ReplicaSet)
		if !ok || rs == nil {
			return nil
		}
		return []string{strconv.Itoa(int(rs.Status.Replicas))}
	}); err != nil {
		return err
	}
	return nil
}
