// Package velero Velero 관련 비즈니스 로직을 관리합니다.
package velero

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstallResult : 설치 결과
type InstallResult struct {
	Status           string                 `json:"status"`
	Message          string                 `json:"message"`
	VeleroNamespace  string                 `json:"velero_namespace"`
	MinioConnected   bool                   `json:"minio_connected"`
	BackupLocation   string                 `json:"backup_location"`
	InstallationTime time.Duration          `json:"installation_time"`
	Details          map[string]interface{} `json:"details,omitempty"`
}

// VeleroStatus : Velero 상태 정보
type VeleroStatus struct {
	PodsInstalled    bool   `json:"pods_installed"`
	HelmRelease      bool   `json:"helm_release"`
	ReleaseNamespace string `json:"release_namespace"`
	IsHealthy        bool   `json:"is_healthy"`
	ErrorMessage     string `json:"error_message,omitempty"`
}

// InstallationError : 설치 에러 정보
type InstallationError struct {
	Type        string   `json:"type"` // "helm_conflict", "pod_failed", "timeout"
	Message     string   `json:"message"`
	Details     string   `json:"details"`
	Suggestions []string `json:"suggestions"`
	Commands    []string `json:"commands"` // 해결을 위한 명령어
}

// Error : error 인터페이스 구현
func (e *InstallationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Message, e.Details)
}

// InstallationProgress : 설치 진행 상황
type InstallationProgress struct {
	Step      string `json:"step"`
	Progress  int    `json:"progress"` // 0-100
	Message   string `json:"message"`
	Estimated string `json:"estimated"` // 남은 시간
	CanCancel bool   `json:"can_cancel"`
}

// BSLStatusInfo : BSL 상태 정보
type BSLStatusInfo struct {
	Phase              string   `json:"phase"`
	Message            string   `json:"message"`
	LastValidationTime string   `json:"last_validation_time"`
	MinioEndpoint      string   `json:"minio_endpoint"`
	Bucket             string   `json:"bucket"`
	Region             string   `json:"region"`
	SuggestedActions   []string `json:"suggested_actions"`
}

// Service : Velero 관련 비즈니스 로직
type Service struct {
	*handler.BaseHandler
}

// NewService : 새로운 Velero 서비스 생성
func NewService(base *handler.BaseHandler) *Service {
	return &Service{
		BaseHandler: base,
	}
}

// InstallVeleroWithMinIOInternal : Velero 설치 및 MinIO 연동 설정 (내부 로직)
func (s *Service) InstallVeleroWithMinIOInternal(client client.Client, ctx context.Context, config config.VeleroConfig, namespace string, force bool) (*InstallResult, error) {
	startTime := time.Now()

	result := &InstallResult{
		Status:          "in_progress",
		VeleroNamespace: namespace,
		Details:         make(map[string]interface{}),
	}

	// 1. Velero 상태 종합 확인
	status, err := s.checkVeleroStatus(client, ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("velero status check failed: %w", err)
	}

	// 2. 스마트 설치 전략 적용
	if err := s.smartInstallVelero(client, ctx, namespace, status, result); err != nil {
		return nil, fmt.Errorf("smart velero installation failed: %w", err)
	}

	// 4. MinIO Secret 생성
	if err := s.createMinIOSecret(client, ctx, config.MinioConfig, namespace, force); err != nil {
		return nil, fmt.Errorf("minio secret creation failed: %w", err)
	}
	if force {
		result.Details["minio_secret"] = "recreated"
	} else {
		result.Details["minio_secret"] = "created"
	}

	// 5. BackupStorageLocation 확인 및 생성
	if err := s.checkAndCreateBackupStorageLocation(client, ctx, config.MinioConfig, namespace, force); err != nil {
		return nil, fmt.Errorf("backup storage location check/creation failed: %w", err)
	}
	if force {
		result.Details["backup_location"] = "recreated"
	} else {
		result.Details["backup_location"] = "created"
	}
	result.BackupLocation = fmt.Sprintf("minio://%s", config.MinioConfig.Endpoint)

	// 6. BSL 상태 조회 및 MinIO 연결 검증
	minioConnected, err := s.validateMinIOConnection(client, ctx, config.MinioConfig, namespace)
	if err != nil {
		return nil, fmt.Errorf("minio connection validation failed: %w", err)
	}
	result.Details["minio_validation"] = "completed"
	result.MinioConnected = minioConnected

	result.Status = "success"
	result.Message = "Velero installed and configured successfully"
	result.InstallationTime = time.Since(startTime)

	return result, nil
}

// checkVeleroStatus : Velero 상태 종합 확인
func (s *Service) checkVeleroStatus(client client.Client, ctx context.Context, namespace string) (*VeleroStatus, error) {
	status := &VeleroStatus{
		PodsInstalled:    false,
		HelmRelease:      false,
		ReleaseNamespace: "",
		IsHealthy:        false,
	}

	// 1. Pod 상태 확인
	podsInstalled, err := s.checkPodsInstalled(client, ctx, namespace)
	if err != nil {
		status.ErrorMessage = fmt.Sprintf("pod check failed: %v", err)
		return status, err
	}
	status.PodsInstalled = podsInstalled

	// 2. Helm Release 상태 확인
	helmRelease, releaseNamespace, err := s.checkHelmRelease(client, ctx, "velero")
	if err != nil {
		status.ErrorMessage = fmt.Sprintf("helm release check failed: %v", err)
		return status, err
	}
	status.HelmRelease = helmRelease
	status.ReleaseNamespace = releaseNamespace

	// 3. 전체 건강 상태 확인
	status.IsHealthy = status.PodsInstalled && status.HelmRelease

	return status, nil
}

// checkPodsInstalled : Velero Pod 설치 여부 확인
func (s *Service) checkPodsInstalled(client client.Client, ctx context.Context, namespace string) (bool, error) {
	pods, err := client.Kubernetes().GetPods(ctx, namespace, "")
	if err != nil {
		return false, fmt.Errorf("failed to get pods: %w", err)
	}

	for _, pod := range pods.(*v1.PodList).Items {
		if pod.Labels["component"] == "velero" && pod.Labels["deploy"] == "velero" {
			return true, nil
		}
	}

	return false, nil
}

// checkHelmRelease : Helm Release 존재 여부 확인
func (s *Service) checkHelmRelease(client client.Client, ctx context.Context, releaseName string) (bool, string, error) {
	releases, err := client.Helm().GetCharts(ctx, "")
	if err != nil {
		return false, "", nil // release 조회 실패는 release가 없다고 간주
	}

	for _, release := range releases {
		if release.Name == releaseName {
			return true, release.Namespace, nil
		}
	}

	return false, "", nil
}

// checkVeleroInstallation : Velero 설치 여부 확인 (기존 호환성 유지)
func (s *Service) checkVeleroInstallation(client client.Client, ctx context.Context, namespace string) (bool, error) {
	status, err := s.checkVeleroStatus(client, ctx, namespace)
	if err != nil {
		return false, err
	}
	return status.IsHealthy, nil
}

// smartInstallVelero : 스마트 Velero 설치 전략
func (s *Service) smartInstallVelero(client client.Client, ctx context.Context, namespace string, status *VeleroStatus, result *InstallResult) error {
	switch {
	case status.IsHealthy:
		// 이미 정상 작동 중
		result.Details["velero_installation"] = "already healthy"
		return nil

	case status.PodsInstalled && !status.HelmRelease:
		// Pod는 있지만 Helm Release가 없는 경우 (수동 설치된 경우)
		result.Details["velero_installation"] = "pods exist but no helm release - recreating helm release"
		return s.recreateHelmRelease(client, ctx, namespace, result)

	case !status.PodsInstalled && status.HelmRelease:
		// Helm Release는 있지만 Pod가 없는 경우 (실패한 설치)
		result.Details["velero_installation"] = "helm release exists but no pods - cleaning up and reinstalling"
		return s.cleanupAndReinstall(client, ctx, namespace, result)

	case status.HelmRelease && status.PodsInstalled:
		// 둘 다 있지만 건강하지 않은 경우
		result.Details["velero_installation"] = "both exist but unhealthy - force reinstalling"
		return s.forceReinstall(client, ctx, namespace, result)

	default:
		// 둘 다 없는 경우 - 새로 설치
		result.Details["velero_installation"] = "fresh installation"
		return s.freshInstall(client, ctx, namespace, result)
	}
}

// recreateHelmRelease : Helm Release 재생성
func (s *Service) recreateHelmRelease(client client.Client, ctx context.Context, namespace string, result *InstallResult) error {
	// 기존 Pod는 그대로 두고 Helm Release만 생성
	if err := s.installVeleroViaHelm(client, ctx); err != nil {
		return fmt.Errorf("failed to recreate helm release: %w", err)
	}
	result.Details["velero_installation"] = "helm release recreated"
	return nil
}

// cleanupAndReinstall : 정리 후 재설치
func (s *Service) cleanupAndReinstall(client client.Client, ctx context.Context, namespace string, result *InstallResult) error {
	// 1. 기존 Helm Release 삭제
	if err := s.safeCleanupRelease(client, ctx, "velero"); err != nil {
		return fmt.Errorf("failed to cleanup existing release: %w", err)
	}

	// 2. 새로 설치
	if err := s.installVeleroViaHelm(client, ctx); err != nil {
		return fmt.Errorf("failed to reinstall: %w", err)
	}

	// 3. Velero 준비 대기
	if err := s.waitForVeleroReady(client, ctx, namespace); err != nil {
		return fmt.Errorf("velero readiness check failed: %w", err)
	}

	result.Details["velero_installation"] = "cleaned up and reinstalled"
	return nil
}

// forceReinstall : 강제 재설치
func (s *Service) forceReinstall(client client.Client, ctx context.Context, namespace string, result *InstallResult) error {
	// 1. 강제 정리
	if err := s.forceCleanupRelease(client, ctx, "velero"); err != nil {
		return fmt.Errorf("failed to force cleanup: %w", err)
	}

	// 2. 새로 설치
	if err := s.installVeleroViaHelm(client, ctx); err != nil {
		return fmt.Errorf("failed to force reinstall: %w", err)
	}

	// 3. Velero 준비 대기
	if err := s.waitForVeleroReady(client, ctx, namespace); err != nil {
		return fmt.Errorf("velero readiness check failed: %w", err)
	}

	result.Details["velero_installation"] = "force reinstalled"
	return nil
}

// freshInstall : 새로 설치
func (s *Service) freshInstall(client client.Client, ctx context.Context, namespace string, result *InstallResult) error {
	// 1. Namespace 생성
	if err := s.createNamespace(client, ctx, namespace); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// 2. Helm으로 설치
	if err := s.installVeleroViaHelm(client, ctx); err != nil {
		return fmt.Errorf("failed to install via helm: %w", err)
	}

	// 3. Velero 준비 대기
	if err := s.waitForVeleroReady(client, ctx, namespace); err != nil {
		return fmt.Errorf("velero readiness check failed: %w", err)
	}

	result.Details["velero_installation"] = "fresh installation completed"
	return nil
}

// isVeleroInstalled : Velero 설치 여부 확인 (boolean 반환)
func (s *Service) isVeleroInstalled(client client.Client, ctx context.Context, namespace string) bool {
	installed, _ := s.checkVeleroInstallation(client, ctx, namespace)
	return installed
}

// installVeleroViaHelm : Helm으로 Velero 설치
func (s *Service) installVeleroViaHelm(client client.Client, ctx context.Context) error {
	// Velero Helm chart URL 설정 (고정값)
	chartURL := "https://github.com/vmware-tanzu/helm-charts/releases/download/velero-10.1.2/velero-10.1.2.tgz"
	releaseName := "velero"
	version := "10.1.2"

	// 기존 release가 있는지 확인하고 삭제
	if err := s.cleanupExistingRelease(client, ctx, releaseName); err != nil {
		return fmt.Errorf("failed to cleanup existing release: %w", err)
	}

	// Velero 설치를 위한 values 설정 (최신 Velero Helm chart 구조)
	values := map[string]interface{}{
		"backupStorageLocation": map[string]interface{}{
			"name":     "default",
			"provider": "aws",
			"bucket":   "velero",
			"config": map[string]interface{}{
				"region": "us-west-2",
			},
		},
		"volumeSnapshotLocation": map[string]interface{}{
			"name":     "default",
			"provider": "aws",
			"config": map[string]interface{}{
				"region": "us-west-2",
			},
		},
		"initContainers": []interface{}{
			map[string]interface{}{
				"name":  "velero-plugin-for-aws",
				"image": "velero/velero-plugin-for-aws:v1.8.0",
				"volumeMounts": []interface{}{
					map[string]interface{}{
						"mountPath": "/target",
						"name":      "plugins",
					},
				},
			},
		},
		"credentials": map[string]interface{}{
			"useSecret":  true,
			"secretName": "cloud-credentials",
		},
	}

	// Helm 설치 실행 (자동 복구 포함)
	err := client.Helm().InstallChart(releaseName, chartURL, version, values)
	if err != nil {
		// 자동 복구 시도
		if recoveryErr := s.autoRecovery(client, ctx, err); recoveryErr != nil {
			return fmt.Errorf("failed to install velero chart: %w (recovery failed: %v)", err, recoveryErr)
		}
	}

	return nil
}

// safeCleanupRelease : 안전한 Helm release 정리
func (s *Service) safeCleanupRelease(client client.Client, ctx context.Context, releaseName string) error {
	// 1. Release 존재 확인
	releases, err := client.Helm().GetCharts(ctx, "")
	if err != nil {
		return nil // release 조회 실패는 release가 없다고 간주
	}

	var targetRelease *release.Release
	for _, release := range releases {
		if release.Name == releaseName {
			targetRelease = release
			break
		}
	}

	if targetRelease == nil {
		return nil // release가 없으면 정리할 것도 없음
	}

	// 2. Release 상태 확인
	if targetRelease.Info.Status == "failed" || targetRelease.Info.Status == "pending-upgrade" {
		// 실패하거나 업그레이드 대기 중인 release는 강제 삭제
		return s.forceCleanupRelease(client, ctx, releaseName)
	}

	// 3. 정상 삭제 시도
	if err := client.Helm().UninstallChart(releaseName, targetRelease.Namespace, false); err != nil {
		// 정상 삭제 실패 시 강제 삭제
		return s.forceCleanupRelease(client, ctx, releaseName)
	}

	// 4. 삭제 완료 대기
	return s.waitForReleaseDeletion(client, ctx, releaseName)
}

// forceCleanupRelease : 강제 Release 정리
func (s *Service) forceCleanupRelease(client client.Client, ctx context.Context, releaseName string) error {
	// 강제 삭제 시도
	if err := client.Helm().UninstallChart(releaseName, "", false); err != nil {
		return fmt.Errorf("failed to force uninstall release '%s': %w", releaseName, err)
	}

	// 삭제 완료 대기
	return s.waitForReleaseDeletion(client, ctx, releaseName)
}

// waitForReleaseDeletion : Release 삭제 완료 대기
func (s *Service) waitForReleaseDeletion(client client.Client, ctx context.Context, releaseName string) error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for release '%s' deletion", releaseName)
		case <-ticker.C:
			releases, err := client.Helm().GetCharts(ctx, "")
			if err != nil {
				continue // 조회 실패 시 다음 시도
			}

			// release가 삭제되었는지 확인
			found := false
			for _, release := range releases {
				if release.Name == releaseName {
					found = true
					break
				}
			}

			if !found {
				return nil // 삭제 완료
			}
		}
	}
}

// cleanupExistingRelease : 기존 Helm release 정리 (기존 호환성 유지)
func (s *Service) cleanupExistingRelease(client client.Client, ctx context.Context, releaseName string) error {
	return s.safeCleanupRelease(client, ctx, releaseName)
}

// autoRecovery : 자동 복구 시도
func (s *Service) autoRecovery(client client.Client, ctx context.Context, err error) error {
	errorStr := err.Error()

	if strings.Contains(errorStr, "cannot re-use a name") {
		// Helm release 이름 충돌
		return s.handleHelmConflict(client, ctx, err)
	}

	if strings.Contains(errorStr, "timeout") {
		// 타임아웃 에러
		return s.handleTimeout(client, ctx, err)
	}

	if strings.Contains(errorStr, "failed to install") {
		// 설치 실패
		return s.handleInstallFailure(client, ctx, err)
	}

	return err // 복구할 수 없는 에러
}

// handleHelmConflict : Helm 충돌 처리
func (s *Service) handleHelmConflict(client client.Client, ctx context.Context, originalErr error) error {
	// 강제 정리 후 재시도
	if err := s.forceCleanupRelease(client, ctx, "velero"); err != nil {
		return s.buildInstallationError("helm_conflict",
			"Helm release name conflict",
			originalErr.Error(),
			[]string{"Try running with force=true", "Manually delete existing release"},
			[]string{"helm uninstall velero", "kubectl delete namespace velero"})
	}

	// 재시도
	return s.installVeleroViaHelm(client, ctx)
}

// handleTimeout : 타임아웃 처리
func (s *Service) handleTimeout(client client.Client, ctx context.Context, originalErr error) error {
	// 더 긴 타임아웃으로 재시도
	extendedCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	return s.installVeleroViaHelm(client, extendedCtx)
}

// handleInstallFailure : 설치 실패 처리
func (s *Service) handleInstallFailure(client client.Client, ctx context.Context, originalErr error) error {
	// 정리 후 재시도
	if err := s.safeCleanupRelease(client, ctx, "velero"); err != nil {
		return s.buildInstallationError("install_failure",
			"Velero installation failed",
			originalErr.Error(),
			[]string{"Check cluster resources", "Verify Helm chart compatibility"},
			[]string{"kubectl get pods -n velero", "helm list"})
	}

	return s.installVeleroViaHelm(client, ctx)
}

// buildInstallationError : 설치 에러 정보 생성
func (s *Service) buildInstallationError(errorType, message, details string, suggestions, commands []string) error {
	return &InstallationError{
		Type:        errorType,
		Message:     message,
		Details:     details,
		Suggestions: suggestions,
		Commands:    commands,
	}
}

// reportProgress : 설치 진행 상황 보고
func (s *Service) reportProgress(step string, progress int, message string, estimated string) *InstallationProgress {
	return &InstallationProgress{
		Step:      step,
		Progress:  progress,
		Message:   message,
		Estimated: estimated,
		CanCancel: progress < 90, // 90% 이하에서만 취소 가능
	}
}

// logInstallationStep : 설치 단계 로깅
func (s *Service) logInstallationStep(step string, message string, details map[string]interface{}) {
	// 실제 구현에서는 로거를 사용
	fmt.Printf("[%s] %s\n", step, message)
	if details != nil {
		for key, value := range details {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
}

// createNamespace : Kubernetes namespace 생성
func (s *Service) createNamespace(client client.Client, ctx context.Context, namespace string) error {
	// Namespace가 이미 존재하는지 확인
	_, err := client.Kubernetes().GetNamespace(ctx, namespace)
	if err == nil {
		return nil // 이미 존재함
	}

	// Namespace 생성
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	// Kubernetes client의 CreateNamespace 메서드 사용
	_, err = client.Kubernetes().CreateNamespace(ctx, ns)
	if err != nil {
		// 이미 존재하는 경우는 성공으로 처리
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create namespace '%s': %w", namespace, err)
	}

	return nil
}

// waitForVeleroReady : Velero가 준비될 때까지 대기
func (s *Service) waitForVeleroReady(client client.Client, ctx context.Context, namespace string) error {
	timeout := time.After(5 * time.Minute) // 5분 타임아웃
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for velero pods to be ready in namespace %s", namespace)
		case <-ticker.C:
			pods, err := client.Kubernetes().GetPods(ctx, namespace, "")
			if err != nil {
				continue // 다음 시도
			}

			allReady := true
			veleroPodsFound := false
			for _, pod := range pods.(*v1.PodList).Items {
				if pod.Labels["component"] == "velero" && pod.Labels["deploy"] == "velero" {
					veleroPodsFound = true
					if pod.Status.Phase != v1.PodRunning {
						allReady = false
						break
					}
				}
			}

			if veleroPodsFound && allReady {
				return nil // Velero Pods are ready
			}
		}
	}
}

// createMinIOSecret : MinIO 연결을 위한 Secret 생성
func (s *Service) createMinIOSecret(client client.Client, ctx context.Context, config config.MinioConfig, namespace string, force bool) error {
	secretName := "cloud-credentials"

	// 기존 Secret 존재 여부 확인
	_, err := client.Kubernetes().GetSecrets(ctx, namespace, secretName)
	if err == nil && !force {
		// 이미 존재하고 force가 false면 스킵
		return nil
	}

	// force=true이고 Secret이 존재하면 삭제 후 재생성
	if err == nil && force {
		// Secret 삭제
		err = client.Kubernetes().DeleteSecret(ctx, namespace, secretName)
		if err != nil {
			return fmt.Errorf("failed to delete existing secret '%s': %w", secretName, err)
		}
	}

	// Secret이 없거나 다른 에러인 경우 생성 시도
	secretData := map[string]string{
		"cloud": fmt.Sprintf(`[default]
aws_access_key_id=%s
aws_secret_access_key=%s
region=us-west-2
`, config.AccessKey, config.SecretKey),
	}

	// Secret 생성
	_, err = client.Kubernetes().CreateSecret(ctx, namespace, secretName, secretData)
	if err != nil {
		// 이미 존재하는 경우 스킵
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create minio secret '%s': %w", secretName, err)
	}

	return nil
}

// checkAndCreateBackupStorageLocation : BSL 확인 및 생성
func (s *Service) checkAndCreateBackupStorageLocation(client client.Client, ctx context.Context, minioConfig config.MinioConfig, namespace string, force bool) error {
	bslName := "minio" // 고정된 BSL 이름

	// 1. BSL 존재 여부 확인
	bsl, err := client.Velero().GetBackupStorageLocation(ctx, namespace, bslName)
	if err != nil {
		// BSL이 없으면 생성
		if strings.Contains(err.Error(), "not found") {
			return s.createBackupStorageLocation(client, ctx, minioConfig, namespace, force)
		}
		return fmt.Errorf("failed to get BackupStorageLocation '%s': %w", bslName, err)
	}

	// 2. BSL PHASE 확인 및 상세 정보 제공
	if bsl.Status.Phase != velerov1.BackupStorageLocationPhaseAvailable {
		// force=true이면 BSL 재생성
		if force {
			return s.createBackupStorageLocation(client, ctx, minioConfig, namespace, force)
		}

		// BSL 상태 정보 수집
		statusInfo := s.buildBSLStatusInfo(bsl, minioConfig)

		// 구체적인 오류 메시지 생성
		errorMsg := s.buildBSLErrorMessage(bslName, string(bsl.Status.Phase), statusInfo)

		return fmt.Errorf("%s", errorMsg)
	}

	return nil
}

// buildBSLStatusInfo : BSL 상태 정보 수집
func (s *Service) buildBSLStatusInfo(bsl *velerov1.BackupStorageLocation, minioConfig config.MinioConfig) BSLStatusInfo {
	statusInfo := BSLStatusInfo{
		Phase:            string(bsl.Status.Phase),
		Message:          bsl.Status.Message,
		MinioEndpoint:    minioConfig.Endpoint,
		Bucket:           "velero",    // 고정값
		Region:           "us-west-2", // 고정값
		SuggestedActions: []string{},
	}

	// LastValidationTime 포맷팅
	if !bsl.Status.LastValidationTime.IsZero() {
		statusInfo.LastValidationTime = bsl.Status.LastValidationTime.Format(time.RFC3339)
	}

	// 오류 메시지 기반 해결 방안 제안
	statusInfo.SuggestedActions = s.suggestBSLActions(bsl.Status.Message, minioConfig)

	return statusInfo
}

// suggestBSLActions : BSL 오류 메시지 기반 해결 방안 제안
func (s *Service) suggestBSLActions(message string, minioConfig config.MinioConfig) []string {
	actions := []string{}

	// 네트워크 연결 문제
	if strings.Contains(message, "no Host in request URL") || strings.Contains(message, "connection refused") {
		actions = append(actions, "Check MinIO endpoint URL format and accessibility")
		actions = append(actions, fmt.Sprintf("Verify MinIO is running at: %s", minioConfig.Endpoint))
	}

	// 인증 문제
	if strings.Contains(message, "access denied") || strings.Contains(message, "invalid credentials") {
		actions = append(actions, "Verify MinIO AccessKey and SecretKey are correct")
		actions = append(actions, "Check MinIO user permissions for the 'velero' bucket")
	}

	// 버킷 접근 문제
	if strings.Contains(message, "bucket") && (strings.Contains(message, "not found") || strings.Contains(message, "does not exist")) {
		actions = append(actions, "Create the 'velero' bucket in MinIO")
		actions = append(actions, "Verify bucket name is 'velero' (case-sensitive)")
	}

	// S3 호환성 문제
	if strings.Contains(message, "S3") && strings.Contains(message, "error") {
		actions = append(actions, "Check MinIO S3 compatibility settings")
		actions = append(actions, "Verify s3ForcePathStyle is enabled")
	}

	// 타임아웃 문제
	if strings.Contains(message, "timeout") || strings.Contains(message, "exceeded maximum number of attempts") {
		actions = append(actions, "Check network connectivity to MinIO")
		actions = append(actions, "Verify MinIO server is responsive")
		actions = append(actions, "Check firewall settings")
	}

	// 기본 해결 방안
	if len(actions) == 0 {
		actions = append(actions, "Check MinIO server status and connectivity")
		actions = append(actions, "Verify MinIO credentials and permissions")
		actions = append(actions, "Check Velero configuration")
	}

	return actions
}

// buildBSLErrorMessage : BSL 오류 메시지 생성
func (s *Service) buildBSLErrorMessage(bslName, phase string, statusInfo BSLStatusInfo) string {
	// 핵심 오류 메시지 (간결하게)
	coreError := s.extractCoreError(statusInfo.Message)

	// 주요 해결 방안만 선별 (최대 3개)
	topActions := s.selectTopActions(statusInfo.SuggestedActions)

	var errorMsg strings.Builder
	errorMsg.WriteString(fmt.Sprintf("BackupStorageLocation '%s' is %s", bslName, strings.ToLower(phase)))

	if coreError != "" {
		errorMsg.WriteString(fmt.Sprintf(": %s", coreError))
	}

	if len(topActions) > 0 {
		errorMsg.WriteString(fmt.Sprintf(" | Try: %s", strings.Join(topActions, ", ")))
	}

	return errorMsg.String()
}

// extractCoreError : 핵심 오류 메시지 추출
func (s *Service) extractCoreError(message string) string {
	if message == "" {
		return ""
	}

	// 가장 중요한 오류 부분만 추출
	if strings.Contains(message, "no Host in request URL") {
		return "Invalid MinIO endpoint URL"
	}
	if strings.Contains(message, "connection refused") {
		return "Cannot connect to MinIO server"
	}
	if strings.Contains(message, "access denied") {
		return "MinIO access denied"
	}
	if strings.Contains(message, "bucket") && strings.Contains(message, "not found") {
		return "MinIO bucket not found"
	}
	if strings.Contains(message, "timeout") || strings.Contains(message, "exceeded maximum number of attempts") {
		return "MinIO connection timeout"
	}

	// 기본적으로 첫 번째 문장만 사용
	parts := strings.Split(message, ":")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}

	return message
}

// selectTopActions : 주요 해결 방안 선별 (최대 3개)
func (s *Service) selectTopActions(actions []string) []string {
	if len(actions) <= 3 {
		return actions
	}

	// 우선순위가 높은 액션들 선별
	priorityActions := []string{}

	for _, action := range actions {
		if strings.Contains(action, "endpoint") || strings.Contains(action, "running") {
			priorityActions = append(priorityActions, action)
		}
	}

	// 부족하면 나머지 추가
	if len(priorityActions) < 3 {
		for _, action := range actions {
			if !contains(priorityActions, action) {
				priorityActions = append(priorityActions, action)
				if len(priorityActions) >= 3 {
					break
				}
			}
		}
	}

	return priorityActions[:3]
}

// contains : 슬라이스에 문자열 포함 여부 확인
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// createBackupStorageLocation : BackupStorageLocation 생성
func (s *Service) createBackupStorageLocation(client client.Client, ctx context.Context, minioConfig config.MinioConfig, namespace string, force bool) error {
	bslName := "minio"

	// 기존 BSL 존재 여부 재확인 (동시성 고려)
	_, err := client.Velero().GetBackupStorageLocation(ctx, namespace, bslName)
	if err == nil && !force {
		// 이미 존재하고 force가 false면 스킵
		return nil
	}

	// force=true이고 BSL이 존재하면 삭제 후 재생성
	if err == nil && force {
		// BSL 삭제
		err = client.Velero().DeleteBackupStorageLocation(ctx, namespace, bslName)
		if err != nil {
			return fmt.Errorf("failed to delete existing BackupStorageLocation '%s': %w", bslName, err)
		}
	}

	bsl := &velerov1.BackupStorageLocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bslName,
			Namespace: namespace,
		},
		Spec: velerov1.BackupStorageLocationSpec{
			Provider: "aws", // MinIO를 AWS S3 호환으로 사용
			StorageType: velerov1.StorageType{
				ObjectStorage: &velerov1.ObjectStorageLocation{
					Bucket: "velero", // 기본 버킷명 사용
					Prefix: "backups",
				},
			},
			Config: map[string]string{
				"region":           "us-west-2", // 기본 리전 사용
				"s3Url":            fmt.Sprintf("http://%s", minioConfig.Endpoint),
				"s3ForcePathStyle": "true",
			},
		},
	}

	err = client.Velero().CreateBackupStorageLocation(ctx, namespace, bsl)
	if err != nil {
		// 이미 존재하는 경우 스킵
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create BackupStorageLocation '%s': %w", bslName, err)
	}

	return nil
}

// validateMinIOConnection : MinIO 연결 및 AccessMode 검증
func (s *Service) validateMinIOConnection(client client.Client, ctx context.Context, minioConfig config.MinioConfig, namespace string) (bool, error) {
	// 1. MinIO 연결 테스트
	_, err := client.Minio().ListBuckets(ctx)
	if err != nil {
		return false, fmt.Errorf("minio connection failed: %w. Please check MinIO endpoint and credentials", err)
	}

	// 2. BackupStorageLocation 상태 확인
	bsl, err := client.Velero().GetBackupStorageLocation(ctx, namespace, "minio")
	if err != nil {
		return false, fmt.Errorf("failed to get BackupStorageLocation 'minio': %w", err)
	}

	if bsl.Status.Phase == velerov1.BackupStorageLocationPhaseAvailable {
		return true, nil
	}

	return false, fmt.Errorf("backup storage location 'minio' is not available. Current phase: %s. Please check MinIO connection and credentials", bsl.Status.Phase)
}

// GetBackupsInternal : Velero 백업 목록 조회 (내부 로직)
func (s *Service) GetBackupsInternal(client client.Client, ctx context.Context, namespace string) (interface{}, error) {
	// Velero 백업 목록 조회
	backups, err := client.Velero().GetBackups(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get backups: %w", err)
	}

	return backups, nil
}

// GetRestoresInternal : Velero 복원 목록 조회 (내부 로직)
func (s *Service) GetRestoresInternal(client client.Client, ctx context.Context, namespace string) (interface{}, error) {
	// Velero 복원 목록 조회
	restores, err := client.Velero().GetRestores(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get restores: %w", err)
	}

	return restores, nil
}
