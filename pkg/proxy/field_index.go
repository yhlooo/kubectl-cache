package proxy

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IndexFields 设置字段索引
func IndexFields(ctx context.Context, c cache.Cache) error {
	if err := IndexFieldsForCoreV1Pod(ctx, c); err != nil {
		return err
	}
	return nil
}

// IndexFieldsForCoreV1Pod 为 corev1.Pod 设置字段索引
func IndexFieldsForCoreV1Pod(ctx context.Context, c cache.Cache) error {
	if err := c.IndexField(ctx, &corev1.Pod{}, "status.phase", func(obj client.Object) []string {
		if pod, ok := obj.(*corev1.Pod); ok && pod != nil {
			return []string{string(pod.Status.Phase)}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
