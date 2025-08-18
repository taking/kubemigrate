package interfaces

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type KubeService interface {
	// Velero 백업 객체 목록을 조회합니다.
	GetResources(ctx context.Context, resource schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error)
}
