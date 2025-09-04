package health

import (
	"context"
	"fmt"

	"taking.kr/velero/pkg/client"
	"taking.kr/velero/pkg/models"
)

// KubernetesHealthChecker :  Kubernetes 연결 상태를 확인하는 헬스체커
type KubernetesHealthChecker struct {
	config models.KubeConfig
}

// NewKubernetesHealthChecker :  새로운 Kubernetes 헬스체커를 생성합니다
func NewKubernetesHealthChecker(config models.KubeConfig) *KubernetesHealthChecker {
	return &KubernetesHealthChecker{
		config: config,
	}
}

func (k *KubernetesHealthChecker) Name() string {
	return "kubernetes"
}

func (k *KubernetesHealthChecker) Check(ctx context.Context) error {
	kubeClient, err := client.NewKubernetesClient(k.config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return kubeClient.HealthCheck(ctx)
}

// VeleroHealthChecker Velero 연결 상태를 확인하는 헬스체커
type VeleroHealthChecker struct {
	config models.KubeConfig
}

// NewVeleroHealthChecker 새로운 Velero 헬스체커를 생성합니다
func NewVeleroHealthChecker(config models.KubeConfig) *VeleroHealthChecker {
	return &VeleroHealthChecker{
		config: config,
	}
}

func (v *VeleroHealthChecker) Name() string {
	return "velero"
}

func (v *VeleroHealthChecker) Check(ctx context.Context) error {
	veleroClient, err := client.NewVeleroClient(v.config)
	if err != nil {
		return fmt.Errorf("failed to create velero client: %w", err)
	}

	return veleroClient.HealthCheck(ctx)
}

// MinioHealthChecker MinIO 연결 상태를 확인하는 헬스체커
type MinioHealthChecker struct {
	config models.MinioConfig
}

// NewMinioHealthChecker 새로운 MinIO 헬스체커를 생성합니다
func NewMinioHealthChecker(config models.MinioConfig) *MinioHealthChecker {
	return &MinioHealthChecker{
		config: config,
	}
}

func (m *MinioHealthChecker) Name() string {
	return "minio"
}

func (m *MinioHealthChecker) Check(ctx context.Context) error {
	minioClient, err := client.NewMinioClient(m.config)
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}

	return minioClient.HealthCheck(ctx)
}

// HelmHealthChecker Helm 연결 상태를 확인하는 헬스체커
type HelmHealthChecker struct {
	config models.KubeConfig
}

// NewHelmHealthChecker 새로운 Helm 헬스체커를 생성합니다
func NewHelmHealthChecker(config models.KubeConfig) *HelmHealthChecker {
	return &HelmHealthChecker{
		config: config,
	}
}

func (h *HelmHealthChecker) Name() string {
	return "helm"
}

func (h *HelmHealthChecker) Check(ctx context.Context) error {
	helmClient, err := client.NewHelmClient(h.config)
	if err != nil {
		return fmt.Errorf("failed to create helm client: %w", err)
	}

	return helmClient.HealthCheck(ctx)
}
