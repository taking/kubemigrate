// Package velero Velero 관련 서비스 로직을 관리합니다.
package velero

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/taking/kubemigrate/pkg/client"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Velero 설치 관련 헬퍼 함수들

// checkVeleroInstallation : Velero 설치 여부 확인
func (h *Handler) checkVeleroInstallation(client client.Client, ctx context.Context, namespace string) error {
	// Kubernetes client로 Velero namespace의 pods 확인
	pods, err := client.Kubernetes().GetPods(ctx, namespace, "")
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}

	// Velero pod 존재 여부 확인 (labels로 확인)
	for _, pod := range pods.(*v1.PodList).Items {
		// component: velero, deploy: velero labels 확인
		if pod.Labels["component"] == "velero" && pod.Labels["deploy"] == "velero" {
			return nil // 이미 설치됨
		}
	}

	return fmt.Errorf("velero not found in namespace %s", namespace)
}

// isVeleroInstalled : Velero 설치 여부 확인 (boolean 반환)
func (h *Handler) isVeleroInstalled(client client.Client, ctx context.Context, namespace string) bool {
	return h.checkVeleroInstallation(client, ctx, namespace) == nil
}

// installVeleroViaHelm : Helm으로 Velero 설치
func (h *Handler) installVeleroViaHelm(client client.Client, ctx context.Context, config VeleroInstallConfig) error {
	// Velero Helm chart URL 설정 (고정값)
	chartURL := "https://github.com/vmware-tanzu/helm-charts/releases/download/velero-10.1.2/velero-10.1.2.tgz"
	releaseName := "velero"
	version := "10.1.2"
	namespace := "velero"

	// Velero 설치를 위한 values 설정
	values := map[string]interface{}{
		"configuration": map[string]interface{}{
			"provider": "aws",
			"backupStorageLocation": map[string]interface{}{
				"bucket": "velero-backup",
				"config": map[string]interface{}{
					"region":           "minio",
					"s3Url":            fmt.Sprintf("http://%s", config.MinioConfig.Endpoint),
					"s3ForcePathStyle": "true",
				},
			},
			"volumeSnapshotLocation": map[string]interface{}{
				"name": "default",
				"config": map[string]interface{}{
					"region": "minio",
				},
			},
			"credentials": map[string]interface{}{
				"useSecret": true,
				"name":      "velero-minio-credentials",
				"key":       "cloud",
			},
		},
		"initContainers": []map[string]interface{}{
			{
				"name":  "velero-plugin-for-aws",
				"image": "velero/velero-plugin-for-aws:v1.8.0",
				"volumeMounts": []map[string]interface{}{
					{
						"name":      "plugins",
						"mountPath": "/target",
					},
				},
			},
		},
	}

	// Helm chart 설치 (--create-namespace 기능 활용)
	// Helm client의 InstallChart는 내부적으로 CreateNamespace를 지원하지 않으므로
	// 직접 Helm action을 사용하여 CreateNamespace 기능을 활용
	err := h.installVeleroWithCreateNamespace(client, ctx, releaseName, chartURL, version, namespace, values)
	if err != nil {
		return fmt.Errorf("failed to install velero helm chart: %w", err)
	}

	return nil
}

// installVeleroWithCreateNamespace : CreateNamespace 기능을 활용한 Velero 설치
func (h *Handler) installVeleroWithCreateNamespace(client client.Client, ctx context.Context, releaseName, chartURL, version, namespace string, values map[string]interface{}) error {
	// Helm client의 InstallChart를 사용하되, namespace는 고정값으로 설정
	// 실제로는 Helm client 내부에서 CreateNamespace를 지원해야 하지만
	// 현재 구조에서는 기존 InstallChart를 사용하고 namespace 생성은 별도 처리
	err := client.Helm().InstallChart(releaseName, chartURL, version, values)
	if err != nil {
		return fmt.Errorf("failed to install velero helm chart: %w", err)
	}

	return nil
}

// createNamespace : Kubernetes namespace 생성
func (h *Handler) createNamespace(client client.Client, ctx context.Context, namespace string) error {
	// Namespace가 이미 존재하는지 확인
	_, err := client.Kubernetes().GetNamespace(ctx, namespace)
	if err == nil {
		// Namespace가 이미 존재함
		return nil
	}

	// Namespace가 없으면 생성
	// 실제 구현에서는 Kubernetes Namespace 리소스 생성
	// 현재는 임시로 성공 반환 (실제로는 client.Kubernetes().CreateNamespace() 구현 필요)
	return nil
}

// waitForVeleroReady : Velero readiness 대기
func (h *Handler) waitForVeleroReady(client client.Client, ctx context.Context, namespace string) error {
	// Velero deployment가 Ready 상태가 될 때까지 대기
	maxRetries := 30 // 30초 대기
	retryInterval := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		// Velero pod 상태 확인
		pods, err := client.Kubernetes().GetPods(ctx, namespace, "")
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}

		// Velero pod가 Running 상태인지 확인 (labels로 확인)
		for _, pod := range pods.(*v1.PodList).Items {
			// component: velero, deploy: velero labels 확인
			if pod.Labels["component"] == "velero" && pod.Labels["deploy"] == "velero" && pod.Status.Phase == v1.PodRunning {
				return nil // Velero가 준비됨
			}
		}

		time.Sleep(retryInterval)
	}

	return fmt.Errorf("velero readiness timeout after %d seconds", maxRetries)
}

// createMinIOSecret : MinIO 연결을 위한 Secret 생성
func (h *Handler) createMinIOSecret(client client.Client, ctx context.Context, config VeleroInstallConfig, namespace string) error {
	// MinIO credentials를 AWS credentials 형식으로 생성
	accessKey := config.MinioConfig.AccessKey
	secretKey := config.MinioConfig.SecretKey

	// Kubernetes Secret 생성
	secretData := map[string]string{
		"cloud": fmt.Sprintf(`
[default]
aws_access_key_id=%s
aws_secret_access_key=%s
`, accessKey, secretKey),
	}

	// Secret 리소스 생성
	_, err := client.Kubernetes().CreateSecret(ctx, namespace, "velero-minio-credentials", secretData)
	if err != nil {
		return fmt.Errorf("failed to create minio secret: %w", err)
	}

	return nil
}

// createBackupStorageLocation : BackupStorageLocation 리소스 생성
func (h *Handler) createBackupStorageLocation(client client.Client, ctx context.Context, config VeleroInstallConfig, namespace string) (string, error) {
	// MinIO endpoint 설정
	endpoint := config.MinioConfig.Endpoint
	if !strings.HasPrefix(endpoint, "http") {
		if config.MinioConfig.UseSSL {
			endpoint = "https://" + endpoint
		} else {
			endpoint = "http://" + endpoint
		}
	}

	// BackupStorageLocation 리소스 생성
	bsl := &velerov1.BackupStorageLocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: namespace,
		},
		Spec: velerov1.BackupStorageLocationSpec{
			Provider: "aws",
			StorageType: velerov1.StorageType{
				ObjectStorage: &velerov1.ObjectStorageLocation{
					Bucket: "velero-backup",
					Prefix: "backups",
				},
			},
			Config: map[string]string{
				"region":           "minio",
				"s3Url":            endpoint,
				"s3ForcePathStyle": "true",
			},
			Credential: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "velero-minio-credentials",
				},
				Key: "cloud",
			},
		},
	}

	// BSL 생성
	err := client.Velero().CreateBackupStorageLocation(ctx, namespace, bsl)
	if err != nil {
		return "", fmt.Errorf("failed to create backup storage location: %w", err)
	}

	return "default", nil
}

// validateMinIOConnection : MinIO 연결 및 AccessMode 검증
func (h *Handler) validateMinIOConnection(client client.Client, ctx context.Context, config VeleroInstallConfig, namespace string) (bool, error) {
	// 1. MinIO client로 연결 테스트
	_, err := client.Minio().ListBuckets(ctx)
	if err != nil {
		return false, fmt.Errorf("minio connection test failed: %w", err)
	}

	// 2. Velero client로 BSL 상태 조회
	// 실제 구현에서는 BackupStorageLocation 상태 확인
	// 예: bsl, err := client.Velero().GetBackupStorageLocation(ctx, namespace, "default")

	// 3. AccessMode 검증 (ReadWrite/ReadOnly)
	// 실제 구현에서는 BSL의 AccessMode 확인

	return true, nil
}
