package services

import (
	"context"
	"taking.kr/velero/interfaces"
)

type KubernetesService struct {
	client interfaces.KubernetesClient
}

func NewKubernetesService(client interfaces.KubernetesClient) *KubernetesService {
	return &KubernetesService{client: client}
}

// GetPods : Pod 목록 반환
func (s *KubernetesService) GetPods(ctx context.Context) (interface{}, error) {
	pods, err := s.client.GetPods(ctx)
	if err != nil {
		return nil, err
	}

	return pods, nil
}

// GetPodsWithStorageClasses : Pod 목록과 클러스터의 StorageClass 목록 함께 반환
func (s *KubernetesService) GetPodsWithStorageClasses(ctx context.Context) (map[string]interface{}, error) {
	pods, err := s.client.GetPods(ctx)
	if err != nil {
		return nil, err
	}

	scs, err := s.client.GetStorageClasses(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"pods":           pods,
		"storageClasses": scs,
	}, nil
}
