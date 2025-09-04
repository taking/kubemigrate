package service

// import (
// 	"context"
// 	"taking.kr/velero/pkg/repository"
// )

// type MigrationService struct {
// 	K8sClient    repository.KubernetesClient
// 	MinioClient  repository.MinioClient
// 	VeleroClient repository.VeleroClient
// 	HelmClient   repository.HelmClient
// }

// // NewMigrationService : 모든 클라이언트를 한 번에 주입
// func NewMigrationService(k8s interfaces.KubernetesClient, minio interfaces.MinioClient, velero interfaces.VeleroClient, helm interfaces.HelmClient) *MigrationService {
// 	return &MigrationService{
// 		K8sClient:    k8s,
// 		MinioClient:  minio,
// 		VeleroClient: velero,
// 		HelmClient:   helm,
// 	}
// }

// // GetPods : Pod 목록 반환
// func (m *MigrationService) GetPods(ctx context.Context) (interface{}, error) {
// 	pods, err := m.K8sClient.GetPods(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return pods, nil
// }

// //
// //// GetPodsWithStorageClasses : Pod 목록과 클러스터의 StorageClass 목록 함께 반환
// //func (s *KubernetesService) GetPodsWithStorageClasses(ctx context.Context) (map[string]interface{}, error) {
// //	pods, err := s.client.GetPods(ctx)
// //	if err != nil {
// //		return nil, err
// //	}
// //
// //	scs, err := s.client.GetStorageClasses(ctx)
// //	if err != nil {
// //		return nil, err
// //	}
// //
// //	return map[string]interface{}{
// //		"pods":           pods,
// //		"storageClasses": scs,
// //	}, nil
// //}
