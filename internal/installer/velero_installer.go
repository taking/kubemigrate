package installer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VeleroInstaller : Velero 설치 담당
type VeleroInstaller struct{}

// NewVeleroInstaller : 새로운 Velero 설치자 생성
func NewVeleroInstaller() *VeleroInstaller {
	return &VeleroInstaller{}
}

// Install : Velero 설치
func (v *VeleroInstaller) Install(ctx context.Context, client client.Client, config VeleroInstallConfig) (*InstallResult, error) {
	start := time.Now()

	result := &InstallResult{
		Status:          "in_progress",
		VeleroNamespace: config.Namespace,
		Force:           config.Force,
		Details:         make(map[string]interface{}),
	}

	// 1. 전략 결정
	strategy, _, err := v.DetermineStrategy(ctx, client, config)
	if err != nil {
		return nil, fmt.Errorf("failed to determine strategy: %w", err)
	}

	// 2. 전략별 실행
	switch strategy {
	case StrategyForceReinstall:
		if err := v.executeForceReinstall(ctx, client, config, result); err != nil {
			return nil, fmt.Errorf("force reinstall failed: %w", err)
		}
	case StrategyFreshInstall:
		if err := v.executeFreshInstall(ctx, client, config, result); err != nil {
			return nil, fmt.Errorf("fresh install failed: %w", err)
		}
	case StrategySkipInstall:
		result.Status = "success"
		result.Message = "Velero is already installed and healthy"
		result.InstallationTime = time.Since(start)
		return result, nil
	}

	// 3. 설치 완료
	result.Status = "success"
	result.Message = "Velero installed successfully"
	result.InstallationTime = time.Since(start)
	return result, nil
}

// Uninstall : Velero 제거
func (v *VeleroInstaller) Uninstall(ctx context.Context, client client.Client, config VeleroUninstallConfig) error {
	// 1. Helm Release 삭제
	if err := v.deleteHelmRelease(ctx, client, "velero", config.Namespace); err != nil {
		return fmt.Errorf("failed to delete helm release: %w", err)
	}

	// 2. Force가 true인 경우 완전 정리
	if config.Force {
		if err := v.performCompleteCleanup(ctx, client, config.Namespace); err != nil {
			return fmt.Errorf("failed to perform complete cleanup: %w", err)
		}
	}

	return nil
}

// Cleanup : Velero 정리
func (v *VeleroInstaller) Cleanup(ctx context.Context, client client.Client, namespace string, force bool) error {
	if force {
		// 완전 정리
		return v.performCompleteCleanup(ctx, client, namespace)
	}

	// 기본 정리 (Helm Release만 삭제)
	return v.deleteHelmRelease(ctx, client, "velero", namespace)
}

// DetermineStrategy : 설치 전략 결정
func (v *VeleroInstaller) DetermineStrategy(ctx context.Context, client client.Client, config VeleroInstallConfig) (InstallationStrategy, *VeleroStatus, error) {
	// Velero 상태 확인
	status, err := v.checkVeleroStatus(ctx, client, config.Namespace)
	if err != nil {
		return "", nil, fmt.Errorf("failed to check velero status: %w", err)
	}

	// Force가 true인 경우 강제 재설치
	if config.Force {
		return StrategyForceReinstall, status, nil
	}

	// Velero가 이미 정상 설치된 경우 스킵
	if status.IsHealthy {
		return StrategySkipInstall, status, nil
	}

	// 그 외의 경우 Fresh 설치
	return StrategyFreshInstall, status, nil
}

// checkVeleroStatus : Velero 상태 확인
func (v *VeleroInstaller) checkVeleroStatus(ctx context.Context, client client.Client, namespace string) (*VeleroStatus, error) {
	// Pods 확인
	podsInstalled, err := v.checkPodsInstalled(ctx, client, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to check pods: %w", err)
	}

	// Helm Release 확인
	helmRelease, err := v.checkHelmRelease(ctx, client, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to check helm release: %w", err)
	}

	status := &VeleroStatus{
		PodsInstalled: podsInstalled,
		HelmRelease:   helmRelease,
		IsHealthy:     podsInstalled && helmRelease,
	}

	return status, nil
}

// checkPodsInstalled : Velero Pods 설치 확인
func (v *VeleroInstaller) checkPodsInstalled(ctx context.Context, client client.Client, namespace string) (bool, error) {
	pods, err := client.Kubernetes().GetPods(ctx, namespace, "")
	if err != nil {
		return false, err
	}

	if podList, ok := pods.(*v1.PodList); ok {
		for _, pod := range podList.Items {
			if strings.Contains(pod.Name, "velero") && pod.Status.Phase == v1.PodRunning {
				return true, nil
			}
		}
	}

	return false, nil
}

// checkHelmRelease : Helm Release 확인
func (v *VeleroInstaller) checkHelmRelease(ctx context.Context, client client.Client, namespace string) (bool, error) {
	installed, _, err := client.Helm().IsChartInstalled("velero")
	if err != nil {
		return false, err
	}
	return installed, nil
}

// executeForceReinstall : 강제 재설치 실행
func (v *VeleroInstaller) executeForceReinstall(ctx context.Context, client client.Client, config VeleroInstallConfig, result *InstallResult) error {
	// 1. 완전 정리 (에러가 있어도 계속 진행)
	fmt.Printf("Starting force cleanup (continuing on errors)...\n")
	v.performCompleteCleanupWithForce(ctx, client, config.Namespace)

	// 2. 정리 완료 후 잠시 대기 (Kubernetes API 안정화)
	fmt.Printf("Waiting for cleanup to stabilize...\n")
	time.Sleep(5 * time.Second)
	fmt.Printf("Cleanup stabilization completed\n")

	// 3. Fresh 설치
	fmt.Printf("Starting fresh installation...\n")
	return v.executeFreshInstall(ctx, client, config, result)
}

// executeFreshInstall : Fresh 설치 실행
func (v *VeleroInstaller) executeFreshInstall(ctx context.Context, client client.Client, config VeleroInstallConfig, result *InstallResult) error {

	// 1. 네임스페이스 확인/생성
	fmt.Printf("  - Ensuring namespace...\n")
	if err := v.ensureNamespaceWithRetry(ctx, client, config.Namespace); err != nil {
		fmt.Printf("    Error: Failed to ensure namespace: %v\n", err)
	} else {
		fmt.Printf("    ✓ Namespace ensured successfully\n")
		// 네임스페이스 생성 후 안정화를 위한 대기
		fmt.Printf("  - Waiting for namespace to be ready...\n")
		time.Sleep(3 * time.Second)
		fmt.Printf("    ✓ Namespace is ready\n")
	}

	// 2. MinIO Secret 생성
	fmt.Printf("  - Creating MinIO Secret...\n")
	if err := v.ensureMinIOSecretWithRetry(ctx, client, config.MinioConfig, config.Namespace); err != nil {
		fmt.Printf("    Error: Failed to ensure minio secret: %v\n", err)
	} else {
		fmt.Printf("    ✓ MinIO Secret created successfully\n")
	}

	// 3. Velero 설치
	fmt.Printf("  - Installing Velero...\n")
	// Helm 설치 전 네임스페이스 존재 재확인
	if err := v.verifyNamespaceExists(ctx, client, config.Namespace); err != nil {
		fmt.Printf("    Error: Namespace verification failed: %v\n", err)
		return fmt.Errorf("namespace verification failed: %w", err)
	}

	if err := v.installVeleroViaHelmWithRetry(ctx, client, config.Namespace, config.MinioConfig); err != nil {
		fmt.Printf("    Error: Failed to install velero: %v\n", err)
		// Velero 설치 실패 시 더 이상 진행할 수 없음
		return fmt.Errorf("critical error: velero installation failed: %w", err)
	} else {
		fmt.Printf("    ✓ Velero installed successfully\n")
	}

	// 4. Velero 준비 대기
	fmt.Printf("  - Waiting for Velero to be ready...\n")
	if err := v.waitForVeleroReady(ctx, client, config.Namespace); err != nil {
		fmt.Printf("    Warning: Velero may not be fully ready: %v\n", err)
	} else {
		fmt.Printf("    ✓ Velero is ready\n")
	}

	// 5. BSL 생성
	fmt.Printf("  - Creating BackupStorageLocation...\n")
	if err := v.ensureBackupStorageLocationWithRetry(ctx, client, config.MinioConfig, config.Namespace); err != nil {
		fmt.Printf("    Warning: Failed to ensure bsl: %v\n", err)
	} else {
		fmt.Printf("    ✓ BackupStorageLocation created successfully\n")
	}

	// 6. MinIO 연결 검증
	fmt.Printf("  - Validating MinIO connection...\n")
	if err := v.validateMinIOConnection(ctx, client, config.MinioConfig, config.Namespace); err != nil {
		fmt.Printf("    Warning: MinIO connection validation failed: %v\n", err)
	} else {
		fmt.Printf("    ✓ MinIO connection validated successfully\n")
	}

	// 설치 결과 요약
	fmt.Printf("Fresh installation completed successfully\n")

	return nil
}

// performCompleteCleanup : 완전 정리
func (v *VeleroInstaller) performCompleteCleanup(ctx context.Context, client client.Client, namespace string) error {
	// 1. Helm Release 삭제
	if err := v.deleteHelmRelease(ctx, client, "velero", namespace); err != nil {
		return fmt.Errorf("failed to delete helm release: %w", err)
	}

	// 2. Helm Release Secrets 삭제
	if err := v.deleteHelmReleaseSecrets(ctx, client, "velero"); err != nil {
		return fmt.Errorf("failed to delete helm release secrets: %w", err)
	}

	// 3. Namespace 삭제 (CRD에 의존하는 리소스들도 함께 삭제)
	if err := v.deleteNamespace(ctx, client, namespace); err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	// 4. Velero CRD 삭제 (클러스터 레벨이므로 별도 삭제 필요)
	if err := v.deleteVeleroCRDs(ctx, client); err != nil {
		return fmt.Errorf("failed to delete velero crds: %w", err)
	}

	return nil
}

// performCompleteCleanupWithForce : 완전 정리 (Force 모드 - 에러가 있어도 계속 진행)
func (v *VeleroInstaller) performCompleteCleanupWithForce(ctx context.Context, client client.Client, namespace string) {

	// 1. Helm Release 삭제
	fmt.Printf("  - Deleting Helm Release...\n")
	if err := v.deleteHelmRelease(ctx, client, "velero", namespace); err != nil {
		fmt.Printf("    Warning: Failed to delete helm release: %v\n", err)
	} else {
		fmt.Printf("    ✓ Helm Release deleted successfully\n")
	}

	// 2. Helm Release Secrets 삭제
	fmt.Printf("  - Deleting Helm Release Secrets...\n")
	if err := v.deleteHelmReleaseSecrets(ctx, client, "velero"); err != nil {
		fmt.Printf("    Warning: Failed to delete helm release secrets: %v\n", err)
	} else {
		fmt.Printf("    ✓ Helm Release Secrets deleted successfully\n")
	}

	// 3. Namespace 삭제
	fmt.Printf("  - Deleting Namespace...\n")
	if err := v.deleteNamespace(ctx, client, namespace); err != nil {
		fmt.Printf("    Warning: Failed to delete namespace: %v\n", err)
	} else {
		fmt.Printf("    ✓ Namespace deleted successfully\n")
	}

	// 4. Velero CRD 삭제
	fmt.Printf("  - Deleting Velero CRDs...\n")
	if err := v.deleteVeleroCRDs(ctx, client); err != nil {
		fmt.Printf("    Warning: Failed to delete velero crds: %v\n", err)
	} else {
		fmt.Printf("    ✓ Velero CRDs deleted successfully\n")
	}

	// 정리 결과 요약
	fmt.Printf("Force cleanup completed successfully\n")
}

// ensureNamespace : 네임스페이스 확인/생성
func (v *VeleroInstaller) ensureNamespace(ctx context.Context, client client.Client, namespace string) error {
	_, err := client.Kubernetes().GetNamespaces(ctx, namespace)
	if err == nil {
		return nil // 이미 존재
	}

	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	_, err = client.Kubernetes().CreateNamespace(ctx, ns)
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	return nil
}

// ensureMinIOSecret : MinIO Secret 확인/생성
func (v *VeleroInstaller) ensureMinIOSecret(ctx context.Context, client client.Client, minioConfig config.MinioConfig, namespace string) error {
	secretName := "cloud-credentials"

	// Secret 존재 확인
	_, err := client.Kubernetes().GetSecrets(ctx, namespace, secretName)
	if err == nil {
		return nil // 이미 존재
	}

	// Secret 생성
	secretData := map[string]string{
		"cloud": fmt.Sprintf(`[default]
aws_access_key_id=%s
aws_secret_access_key=%s
region=minio
`, minioConfig.AccessKey, minioConfig.SecretKey),
	}

	_, err = client.Kubernetes().CreateSecret(ctx, namespace, secretName, secretData)
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// installVeleroViaHelm : Helm을 통한 Velero 설치
func (v *VeleroInstaller) installVeleroViaHelm(ctx context.Context, client client.Client, namespace string, minioConfig config.MinioConfig) error {
	chartURL := "https://github.com/vmware-tanzu/helm-charts/releases/download/velero-11.1.0/velero-11.1.0.tgz"
	releaseName := "velero"
	version := "11.1.0"

	values := map[string]interface{}{
		"image": map[string]interface{}{
			"repository": "docker.io/velero/velero",
			"tag":        "v1.17.0",
		},
		"kubectl": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "docker.io/bitnamilegacy/kubectl",
			},
		},
		"credentials": map[string]interface{}{
			"useSecret":      true,
			"existingSecret": "cloud-credentials",
		},
		"features":                 "EnableCSI",
		"deployNodeAgent":          true,
		"defaultVolumesToFsBackup": true,
		"configuration": map[string]interface{}{
			"backupStorageLocation": []interface{}{
				map[string]interface{}{
					"name":     "minio",
					"provider": "aws",
					"bucket":   "velero",
					"config": map[string]interface{}{
						"region":           "minio",
						"s3Url":            fmt.Sprintf("http://%s", minioConfig.Endpoint),
						"s3ForcePathStyle": "true",
					},
					"credential": map[string]interface{}{
						"name": "cloud-credentials",
						"key":  "cloud",
					},
				},
			},
			"volumeSnapshotLocation": []interface{}{
				map[string]interface{}{
					"name":     "minio-snapshot",
					"provider": "aws",
					"config": map[string]interface{}{
						"region": "minio",
					},
					"credential": map[string]interface{}{
						"name": "cloud-credentials",
						"key":  "cloud",
					},
				},
			},
		},
		"initContainers": []interface{}{
			map[string]interface{}{
				"name":            "velero-plugin-for-aws",
				"image":           "docker.io/velero/velero-plugin-for-aws:v1.13.0",
				"imagePullPolicy": "IfNotPresent",
				"volumeMounts": []interface{}{
					map[string]interface{}{
						"name":      "plugins",
						"mountPath": "/target",
					},
				},
			},
		},
	}

	// 1. 먼저 install 시도
	if err := client.Helm().InstallChart(releaseName, chartURL, version, namespace, values); err != nil {
		// 2. install 실패 시 upgrade 시도 (기존 release가 있는 경우)
		if strings.Contains(err.Error(), "cannot re-use a name") {
			if upgradeErr := client.Helm().UpgradeChart(releaseName, chartURL, version, namespace, values); upgradeErr != nil {
				return fmt.Errorf("failed to upgrade velero chart: %w", upgradeErr)
			}
		} else {
			return fmt.Errorf("failed to install velero chart: %w", err)
		}
	}

	return nil
}

// ensureBackupStorageLocation : BSL 확인/생성
func (v *VeleroInstaller) ensureBackupStorageLocation(ctx context.Context, client client.Client, minioConfig config.MinioConfig, namespace string) error {
	bslName := "minio"

	bsl := &velerov1.BackupStorageLocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bslName,
			Namespace: namespace,
		},
		Spec: velerov1.BackupStorageLocationSpec{
			Provider: "aws",
			StorageType: velerov1.StorageType{
				ObjectStorage: &velerov1.ObjectStorageLocation{
					Bucket: "velero",
					Prefix: "backups",
				},
			},
			Config: map[string]string{
				"region":           "us-west-2",
				"s3Url":            fmt.Sprintf("http://%s", minioConfig.Endpoint),
				"s3ForcePathStyle": "true",
			},
			Credential: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "cloud-credentials",
				},
				Key: "cloud",
			},
			Default: true,
		},
	}

	return client.Velero().CreateBackupStorageLocation(ctx, namespace, bsl)
}

// deleteHelmRelease : Helm Release 삭제
func (v *VeleroInstaller) deleteHelmRelease(ctx context.Context, client client.Client, releaseName, namespace string) error {
	// 여러 네임스페이스에서 시도
	namespaces := []string{namespace, "velero", "default"}

	for _, ns := range namespaces {
		if err := client.Helm().UninstallChart(releaseName, ns, false); err == nil {
			return nil // 삭제 성공
		}
		// 삭제 실패해도 다음 네임스페이스 시도
	}

	return fmt.Errorf("failed to uninstall helm release '%s' from any namespace", releaseName)
}

// deleteVeleroCRDs : Velero CRD 삭제
func (v *VeleroInstaller) deleteVeleroCRDs(ctx context.Context, client client.Client) error {
	// Velero CRD 목록 (클러스터 레벨 리소스)
	crdNames := []string{
		"backups.velero.io",
		"backupstoragelocations.velero.io",
		"volumesnapshotlocations.velero.io",
		"restores.velero.io",
		"podvolumebackups.velero.io",
		"podvolumerestores.velero.io",
		"downloadrequests.velero.io",
		"backuprepositories.velero.io",
		"serverstatusrequests.velero.io",
	}

	fmt.Printf("Deleting Velero CRDs (cluster-level resources)...\n")

	var failedCRDs []string
	for _, crdName := range crdNames {
		fmt.Printf("  - Deleting CRD: %s\n", crdName)

		if err := client.Kubernetes().DeleteCRD(ctx, crdName); err != nil {
			fmt.Printf("    Failed to delete CRD %s: %v\n", crdName, err)
			failedCRDs = append(failedCRDs, crdName)
		} else {
			fmt.Printf("    Successfully deleted CRD: %s\n", crdName)
		}
	}

	if len(failedCRDs) > 0 {
		fmt.Printf("Warning: Failed to delete %d CRDs: %s\n", len(failedCRDs), strings.Join(failedCRDs, ", "))
		fmt.Printf("Manual command: kubectl delete crd %s\n", strings.Join(failedCRDs, " "))
		return fmt.Errorf("failed to delete %d CRDs: %s", len(failedCRDs), strings.Join(failedCRDs, ", "))
	}

	fmt.Printf("Successfully deleted all Velero CRDs\n")
	return nil
}

// deleteHelmReleaseSecrets : Helm Release Secrets 삭제
func (v *VeleroInstaller) deleteHelmReleaseSecrets(ctx context.Context, client client.Client, releaseName string) error {
	// 모든 네임스페이스에서 Helm Release Secrets 조회
	namespaces := []string{"velero", "default", "kube-system"}

	for _, namespace := range namespaces {
		secrets, err := client.Kubernetes().GetSecrets(ctx, namespace, "")
		if err != nil {
			continue // 조회 실패해도 다음 네임스페이스 시도
		}

		if secretList, ok := secrets.([]v1.Secret); ok {
			for _, secret := range secretList {
				// Helm Release Secret 패턴 확인
				if strings.HasPrefix(secret.Name, fmt.Sprintf("sh.helm.release.v1.%s.", releaseName)) {
					if err := client.Kubernetes().DeleteSecret(ctx, namespace, secret.Name); err != nil {
						fmt.Printf("Failed to delete secret %s in namespace %s: %v\n", secret.Name, namespace, err)
					} else {
						fmt.Printf("Deleted secret %s in namespace %s\n", secret.Name, namespace)
					}
				}
			}
		}
	}

	return nil
}

// deleteNamespace : 네임스페이스 삭제
func (v *VeleroInstaller) deleteNamespace(ctx context.Context, client client.Client, namespace string) error {
	// 네임스페이스 존재 확인
	_, err := client.Kubernetes().GetNamespaces(ctx, namespace)
	if err != nil {
		fmt.Printf("Namespace '%s' does not exist, skipping deletion\n", namespace)
		return nil // 이미 존재하지 않음
	}

	fmt.Printf("Deleting namespace: %s\n", namespace)

	// 네임스페이스 삭제
	if err := client.Kubernetes().DeleteNamespace(ctx, namespace); err != nil {
		return fmt.Errorf("failed to delete namespace '%s': %w", namespace, err)
	}

	fmt.Printf("Successfully deleted namespace: %s\n", namespace)
	return nil
}

// verifyNamespaceExists : 네임스페이스 존재 확인 (재시도 포함)
func (v *VeleroInstaller) verifyNamespaceExists(ctx context.Context, client client.Client, namespace string) error {
	return v.retryOperation(func() error {
		_, err := client.Kubernetes().GetNamespaces(ctx, namespace)
		if err != nil {
			return fmt.Errorf("namespace '%s' does not exist: %w", namespace, err)
		}
		return nil
	}, 5, 2*time.Second)
}

// retryOperation : 재시도 로직을 포함한 작업 실행
func (v *VeleroInstaller) retryOperation(operation func() error, maxRetries int, delay time.Duration) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := operation(); err != nil {
			lastErr = err
			if i < maxRetries-1 {
				time.Sleep(delay)
				continue
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}

// ensureNamespaceWithRetry : 네임스페이스 확인/생성 (재시도 포함)
func (v *VeleroInstaller) ensureNamespaceWithRetry(ctx context.Context, client client.Client, namespace string) error {
	return v.retryOperation(func() error {
		return v.ensureNamespace(ctx, client, namespace)
	}, 3, 2*time.Second)
}

// ensureMinIOSecretWithRetry : MinIO Secret 확인/생성 (재시도 포함)
func (v *VeleroInstaller) ensureMinIOSecretWithRetry(ctx context.Context, client client.Client, minioConfig config.MinioConfig, namespace string) error {
	return v.retryOperation(func() error {
		return v.ensureMinIOSecret(ctx, client, minioConfig, namespace)
	}, 3, 2*time.Second)
}

// installVeleroViaHelmWithRetry : Velero Helm 설치 (재시도 포함)
func (v *VeleroInstaller) installVeleroViaHelmWithRetry(ctx context.Context, client client.Client, namespace string, minioConfig config.MinioConfig) error {
	return v.retryOperation(func() error {
		return v.installVeleroViaHelm(ctx, client, namespace, minioConfig)
	}, 3, 5*time.Second)
}

// ensureBackupStorageLocationWithRetry : BSL 확인/생성 (재시도 포함)
func (v *VeleroInstaller) ensureBackupStorageLocationWithRetry(ctx context.Context, client client.Client, minioConfig config.MinioConfig, namespace string) error {
	return v.retryOperation(func() error {
		return v.ensureBackupStorageLocation(ctx, client, minioConfig, namespace)
	}, 3, 2*time.Second)
}

// waitForVeleroReady : Velero Pod 준비 대기
func (v *VeleroInstaller) waitForVeleroReady(ctx context.Context, client client.Client, namespace string) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for velero pods to be ready in namespace %s", namespace)
		case <-ticker.C:
			ready, err := v.checkVeleroPodsReady(ctx, client, namespace)
			if err != nil {
				continue
			}
			if ready {
				return nil
			}
		}
	}
}

// checkVeleroPodsReady : Velero Pods 준비 상태 확인
func (v *VeleroInstaller) checkVeleroPodsReady(ctx context.Context, client client.Client, namespace string) (bool, error) {
	pods, err := client.Kubernetes().GetPods(ctx, namespace, "")
	if err != nil {
		return false, err
	}

	if podList, ok := pods.(*v1.PodList); ok {
		veleroPodsFound := false
		allReady := true

		for _, pod := range podList.Items {
			if strings.Contains(pod.Name, "velero") {
				veleroPodsFound = true
				if pod.Status.Phase != v1.PodRunning {
					allReady = false
					break
				}
			}
		}

		return veleroPodsFound && allReady, nil
	}

	return false, nil
}

// validateMinIOConnection : MinIO 연결 검증
func (v *VeleroInstaller) validateMinIOConnection(ctx context.Context, client client.Client, minioConfig config.MinioConfig, namespace string) error {
	// MinIO 클라이언트 연결 테스트
	if _, err := client.Minio().ListBuckets(ctx); err != nil {
		return fmt.Errorf("minio connection failed: %w", err)
	}

	// BSL 상태 확인
	bsls, err := client.Velero().GetBackupStorageLocations(ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to get backup storage locations: %w", err)
	}

	for _, bsl := range bsls {
		if bsl.Name == "minio" {
			if bsl.Status.Phase != velerov1.BackupStorageLocationPhaseAvailable {
				return fmt.Errorf("backup storage location 'minio' is not available (phase: %s)", bsl.Status.Phase)
			}
			return nil
		}
	}

	return fmt.Errorf("backup storage location 'minio' not found")
}
