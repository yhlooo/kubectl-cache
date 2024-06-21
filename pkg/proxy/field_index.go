package proxy

import (
	"context"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IndexFields 设置字段索引
func IndexFields(ctx context.Context, c cache.Cache) error {
	if err := IndexFieldsForCoreV1Pod(ctx, c); err != nil {
		return err
	}
	if err := IndexFieldsForCoreV1Event(ctx, c); err != nil {
		return err
	}
	if err := IndexFieldsForAppsV1ReplicaSet(ctx, c); err != nil {
		return err
	}
	return nil
}

// IndexFieldsForCoreV1Pod 为 corev1.Pod 设置字段索引
func IndexFieldsForCoreV1Pod(ctx context.Context, c cache.Cache) error {
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
