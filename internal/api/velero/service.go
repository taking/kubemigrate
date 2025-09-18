// Package velero Velero 관련 비즈니스 로직을 관리합니다.
package velero

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/job"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Service : Velero 관련 비즈니스 로직
type Service struct {
	*handler.BaseHandler
	jobManager job.JobManager
}

// NewService : 새로운 Velero 서비스 생성
func NewService(base *handler.BaseHandler) *Service {
	return &Service{
		BaseHandler: base,
		jobManager:  job.NewMemoryJobManager(),
	}
}

// InstallVeleroWithMinIOInternal : Velero 설치 및 MinIO 연동 설정 (비동기)
// 상위 흐름만 담당하고 세부 로직은 헬퍼로 분리합니다.
func (s *Service) InstallVeleroWithMinIOInternal(
	client client.Client,
	ctx context.Context,
	cfg config.VeleroConfig,
	namespace string,
	force bool,
) (interface{}, error) {
	// Job ID 생성
	jobID := fmt.Sprintf("velero-install-%d", time.Now().UnixNano())

	// Job 생성 (민감한 정보 제외)
	metadata := map[string]interface{}{
		"namespace":       namespace,
		"force":           force,
		"minioEndpoint":   cfg.MinioConfig.Endpoint,
		"veleroNamespace": namespace,
	}

	jobInfo := s.jobManager.CreateJob(jobID, metadata)

	// 백그라운드에서 설치 시작
	go s.installVeleroWithMinIOBackground(client, ctx, jobID, cfg, namespace, force)

	// 즉시 응답 반환
	return map[string]interface{}{
		"status":    "processing",
		"jobId":     jobID,
		"message":   "Velero installation started",
		"statusUrl": fmt.Sprintf("/api/v1/velero/status/%s", jobID),
		"logsUrl":   fmt.Sprintf("/api/v1/velero/logs/%s", jobID),
		"job":       jobInfo,
	}, nil
}

// installVeleroWithMinIOBackground : 백그라운드에서 Velero 설치
func (s *Service) installVeleroWithMinIOBackground(
	client client.Client,
	ctx context.Context,
	jobID string,
	cfg config.VeleroConfig,
	namespace string,
	force bool,
) {
	start := time.Now()

	// 백그라운드 작업을 위한 새로운 context 생성 (30분 timeout)
	bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 1. 상태 확인 (재시도 로직 포함)
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 10, "Checking Velero status...")
	s.jobManager.AddJobLog(jobID, "Starting Velero installation process")

	var status *VeleroStatus
	err := s.jobManager.RetryOperation(jobID, "Velero status check", 3, func() error {
		var statusErr error
		status, statusErr = s.checkVeleroStatus(client, bgCtx, namespace)
		return statusErr
	})
	if err != nil {
		s.jobManager.FailJob(jobID, err)
		return
	}
	s.jobManager.AddJobLog(jobID, "Velero status check completed successfully")

	// 2. 설치 전략 결정 및 실행 (재시도 로직 포함)
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 30, "Applying installation strategy...")
	s.jobManager.AddJobLog(jobID, "Determining installation strategy")

	result := &InstallResult{
		Status:          "in_progress",
		VeleroNamespace: namespace,
		Force:           force,
		Details:         make(map[string]interface{}),
	}

	err = s.jobManager.RetryOperation(jobID, "Installation strategy", 3, func() error {
		return s.applyInstallStrategy(client, bgCtx, namespace, status, result)
	})
	if err != nil {
		s.jobManager.FailJob(jobID, err)
		return
	}
	s.jobManager.AddJobLog(jobID, "Installation strategy applied successfully")

	// 3. MinIO Secret 생성/재생성 (재시도 로직 포함)
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 50, "Creating MinIO secret...")
	s.jobManager.AddJobLog(jobID, "Ensuring MinIO secret")

	err = s.jobManager.RetryOperation(jobID, "MinIO secret creation", 3, func() error {
		return s.ensureMinIOSecret(client, bgCtx, cfg.MinioConfig, namespace, force, result)
	})
	if err != nil {
		s.jobManager.FailJob(jobID, err)
		return
	}
	s.jobManager.AddJobLog(jobID, "MinIO secret created successfully")

	// 4. BackupStorageLocation 생성/검증 (재시도 로직 포함)
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 70, "Creating backup storage location...")
	s.jobManager.AddJobLog(jobID, "Ensuring backup storage location")

	err = s.jobManager.RetryOperation(jobID, "Backup storage location creation", 3, func() error {
		return s.ensureBackupStorageLocation(client, bgCtx, cfg.MinioConfig, namespace, force, result)
	})
	if err != nil {
		s.jobManager.FailJob(jobID, err)
		return
	}
	s.jobManager.AddJobLog(jobID, "Backup storage location created successfully")

	// 5. MinIO 연결 및 BSL 상태 검증 (재시도 로직 포함)
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 90, "Validating MinIO connection...")
	s.jobManager.AddJobLog(jobID, "Validating MinIO connection and BSL status")

	var ok bool
	err = s.jobManager.RetryOperation(jobID, "MinIO validation", 3, func() error {
		var validationErr error
		ok, validationErr = s.validateMinIOConnection(client, bgCtx, cfg.MinioConfig, namespace)
		return validationErr
	})
	if err != nil {
		s.jobManager.FailJob(jobID, err)
		return
	}
	result.MinioConnected = ok
	result.Details["minio_validation"] = "completed"
	s.jobManager.AddJobLog(jobID, "MinIO connection validated successfully")

	// 6. 완료
	result.Status = "success"
	result.Message = "Velero installed and configured successfully"
	result.BackupLocation = fmt.Sprintf("minio://%s", cfg.MinioConfig.Endpoint)
	result.InstallationTime = time.Since(start)

	s.jobManager.CompleteJob(jobID, result)
	s.jobManager.AddJobLog(jobID, "Velero installation completed successfully")
}

// GetJobStatusInternal : 작업 상태 조회 (내부 로직)
func (s *Service) GetJobStatusInternal(jobID string) (interface{}, error) {
	job, exists := s.jobManager.GetJob(jobID)
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// GetJobLogsInternal : 작업 로그 조회 (내부 로직)
func (s *Service) GetJobLogsInternal(jobID string) (interface{}, error) {
	job, exists := s.jobManager.GetJob(jobID)
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return map[string]interface{}{
		"jobId":   job.JobID,
		"logs":    job.Logs,
		"status":  job.Status,
		"message": job.Message,
	}, nil
}

// GetAllJobsInternal : 모든 작업 조회 (내부 로직)
func (s *Service) GetAllJobsInternal() (interface{}, error) {
	jobs := s.jobManager.GetAllJobs()
	return map[string]interface{}{
		"jobs":  jobs,
		"count": len(jobs),
	}, nil
}

//
// ─── 상태 확인 로직 ───────────────────────────────────────────────────────────────
//

// checkVeleroStatus : Velero 상태 종합 확인
func (s *Service) checkVeleroStatus(client client.Client, ctx context.Context, namespace string) (*VeleroStatus, error) {
	status := &VeleroStatus{}

	podsInstalled, err := s.checkPodsInstalled(client, ctx, namespace)
	if err != nil {
		status.ErrorMessage = fmt.Sprintf("pod check failed: %v", err)
		return status, err
	}
	status.PodsInstalled = podsInstalled

	helmRelease, releaseNamespace, err := s.checkHelmRelease(client, ctx, "velero")
	if err != nil {
		status.ErrorMessage = fmt.Sprintf("helm release check failed: %v", err)
		// release 조회 실패는 에러지만 진행 가능하도록 false로 표시
	}
	status.HelmRelease = helmRelease
	status.ReleaseNamespace = releaseNamespace

	status.IsHealthy = status.PodsInstalled && status.HelmRelease
	return status, nil
}

// checkPodsInstalled : Velero Pod 설치 여부 확인
func (s *Service) checkPodsInstalled(client client.Client, ctx context.Context, namespace string) (bool, error) {
	raw, err := client.Kubernetes().GetPods(ctx, namespace, "")
	if err != nil {
		return false, fmt.Errorf("failed to get pods: %w", err)
	}

	// client.Kubernetes().GetPods 는 v1.PodList 타입을 반환한다고 가정
	if podList, ok := raw.(*v1.PodList); ok {
		for _, pod := range podList.Items {
			if pod.Labels["component"] == "velero" && pod.Labels["deploy"] == "velero" {
				return true, nil
			}
		}
	}
	return false, nil
}

// checkHelmRelease : Helm Release 존재 여부 확인
func (s *Service) checkHelmRelease(client client.Client, ctx context.Context, releaseName string) (bool, string, error) {
	releases, err := client.Helm().GetCharts(ctx, "")
	if err != nil {
		// Helm 조회 실패는 에러지만 "없음"으로 취급해 상위에서 결정하도록 함
		return false, "", nil
	}
	for _, r := range releases {
		if r.Name == releaseName {
			return true, r.Namespace, nil
		}
	}
	return false, "", nil
}

//
// ─── 설치 전략 관련 ──────────────────────────────────────────────────────────────
//

func (s *Service) applyInstallStrategy(
	client client.Client,
	ctx context.Context,
	namespace string,
	status *VeleroStatus,
	result *InstallResult,
) error {
	switch {
	case status.IsHealthy:
		result.Details["velero_installation"] = "already healthy"
		return nil

	case !status.PodsInstalled && !status.HelmRelease:
		result.Details["velero_installation"] = "fresh installation"
		return s.freshInstall(client, ctx, namespace, result)

	case status.PodsInstalled && !status.HelmRelease:
		result.Details["velero_installation"] = "pods exist but helm release missing"
		return s.recreateHelmRelease(client, ctx, namespace, result)

	case !status.PodsInstalled && status.HelmRelease:
		result.Details["velero_installation"] = "helm release exists but pods missing"
		return s.cleanupAndReinstall(client, ctx, namespace, result)

	default:
		result.Details["velero_installation"] = "unhealthy - force reinstall"
		return s.forceReinstall(client, ctx, namespace, result)
	}
}

// freshInstall : 새로 설치
func (s *Service) freshInstall(client client.Client, ctx context.Context, namespace string, result *InstallResult) error {
	// 1. Helm으로 설치 (--create-namespace 옵션으로 네임스페이스 자동 생성)
	if err := s.installVeleroViaHelm(client, ctx, namespace); err != nil {
		return fmt.Errorf("failed to install via helm: %w", err)
	}
	// 2. 준비 대기
	if err := s.waitForVeleroReady(client, ctx, namespace); err != nil {
		return fmt.Errorf("velero readiness check failed: %w", err)
	}
	result.Details["velero_installation"] = "fresh installation completed"
	return nil
}

// recreateHelmRelease : Pod는 두고 Helm Release만 재생성
func (s *Service) recreateHelmRelease(client client.Client, ctx context.Context, namespace string, result *InstallResult) error {
	if err := s.installVeleroViaHelm(client, ctx, namespace); err != nil {
		return fmt.Errorf("failed to recreate helm release: %w", err)
	}
	result.Details["velero_installation"] = "helm release recreated"
	return nil
}

// cleanupAndReinstall : 정리 후 재설치
func (s *Service) cleanupAndReinstall(client client.Client, ctx context.Context, namespace string, result *InstallResult) error {
	// safe cleanup
	if err := s.safeCleanupRelease(client, ctx, "velero"); err != nil {
		return fmt.Errorf("failed to cleanup existing release: %w", err)
	}
	if err := s.installVeleroViaHelm(client, ctx, namespace); err != nil {
		return fmt.Errorf("failed to reinstall: %w", err)
	}
	if err := s.waitForVeleroReady(client, ctx, namespace); err != nil {
		return fmt.Errorf("velero readiness check failed: %w", err)
	}
	result.Details["velero_installation"] = "cleaned up and reinstalled"
	return nil
}

// forceReinstall : 강제 재설치
func (s *Service) forceReinstall(client client.Client, ctx context.Context, namespace string, result *InstallResult) error {
	// force cleanup across namespaces
	if err := s.forceCleanupRelease(client, ctx, "velero", namespace); err != nil {
		return fmt.Errorf("failed to force cleanup: %w", err)
	}
	if err := s.installVeleroViaHelm(client, ctx, namespace); err != nil {
		return fmt.Errorf("failed to force reinstall: %w", err)
	}
	if err := s.waitForVeleroReady(client, ctx, namespace); err != nil {
		return fmt.Errorf("velero readiness check failed: %w", err)
	}
	result.Details["velero_installation"] = "force reinstalled"
	return nil
}

//
// ─── Helm install / cleanup / wait helpers ────────────────────────────────────────
//

func (s *Service) installVeleroViaHelm(client client.Client, ctx context.Context, namespace string) error {
	chartURL := "https://github.com/vmware-tanzu/helm-charts/releases/download/velero-10.1.2/velero-10.1.2.tgz"
	releaseName := "velero"
	version := "10.1.2"

	// 간단한 values 설정 (최소한의 설정만)
	values := map[string]interface{}{
		"credentials": map[string]interface{}{
			"useSecret":  true,
			"secretName": "cloud-credentials",
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

// forceUninstallExistingRelease : 기존 Helm release 강제 삭제
func (s *Service) forceUninstallExistingRelease(client client.Client, ctx context.Context, releaseName, namespace string) error {
	// 1. 모든 네임스페이스에서 release 조회
	releases, err := client.Helm().GetCharts(ctx, "")
	if err != nil {
		// 조회 실패는 release가 없다고 간주
		return nil
	}

	// 2. 해당 release가 있는지 확인
	var targetRelease *release.Release
	for _, r := range releases {
		if r.Name == releaseName {
			targetRelease = r
			break
		}
	}

	// 3. release가 없으면 그냥 진행
	if targetRelease == nil {
		return nil
	}

	// 4. release 삭제 시도 (여러 네임스페이스에서 시도)
	uninstallNamespaces := []string{namespace, "velero", "default"}

	for _, ns := range uninstallNamespaces {
		if err := client.Helm().UninstallChart(releaseName, ns, false); err == nil {
			// 삭제 성공
			return nil
		}
		// 삭제 실패해도 다음 네임스페이스 시도
	}

	// 5. 모든 시도 실패 시 에러 반환
	return fmt.Errorf("failed to uninstall existing release '%s' from any namespace", releaseName)
}

func (s *Service) safeCleanupRelease(client client.Client, ctx context.Context, releaseName string) error {
	releases, err := client.Helm().GetCharts(ctx, "")
	if err != nil {
		return nil
	}
	var target *release.Release
	for _, r := range releases {
		if r.Name == releaseName {
			target = r
			break
		}
	}
	if target == nil {
		return nil
	}
	if target.Info.Status == "failed" || target.Info.Status == "pending-upgrade" {
		return s.forceCleanupRelease(client, ctx, releaseName, target.Namespace)
	}
	if err := client.Helm().UninstallChart(releaseName, target.Namespace, false); err != nil {
		return s.forceCleanupRelease(client, ctx, releaseName, target.Namespace)
	}
	return s.waitForReleaseDeletion(client, ctx, releaseName)
}

func (s *Service) forceCleanupRelease(client client.Client, ctx context.Context, releaseName, namespace string) error {
	if err := client.Helm().UninstallChart(releaseName, namespace, false); err != nil {
		return fmt.Errorf("failed to force uninstall release '%s' in namespace '%s': %w", releaseName, namespace, err)
	}
	return s.waitForReleaseDeletion(client, ctx, releaseName)
}

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
				continue
			}
			found := false
			for _, r := range releases {
				if r.Name == releaseName {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}
	}
}

func (s *Service) enhancedCleanupRelease(client client.Client, ctx context.Context, releaseName string) error {
	releases, err := client.Helm().GetCharts(ctx, "")
	if err != nil {
		return nil
	}
	var target *release.Release
	for _, r := range releases {
		if r.Name == releaseName {
			target = r
			break
		}
	}
	if target == nil {
		return nil
	}
	switch target.Info.Status {
	case "failed", "pending-upgrade", "pending-install":
		if err := s.forceCleanupRelease(client, ctx, releaseName, target.Namespace); err != nil {
			return fmt.Errorf("failed to force cleanup release '%s' in namespace '%s': %w", releaseName, target.Namespace, err)
		}
	case "deployed", "superseded":
		if err := client.Helm().UninstallChart(releaseName, target.Namespace, false); err != nil {
			if err := s.forceCleanupRelease(client, ctx, releaseName, target.Namespace); err != nil {
				return fmt.Errorf("failed to cleanup release '%s' in namespace '%s': %w", releaseName, target.Namespace, err)
			}
		}
	default:
		if err := s.forceCleanupRelease(client, ctx, releaseName, target.Namespace); err != nil {
			return fmt.Errorf("failed to force cleanup release '%s' in namespace '%s': %w", releaseName, target.Namespace, err)
		}
	}
	return s.waitForReleaseDeletion(client, ctx, releaseName)
}

func (s *Service) autoRecovery(client client.Client, ctx context.Context, err error, namespace string) error {
	errStr := err.Error()
	if strings.Contains(errStr, "cannot re-use a name") {
		return s.handleHelmConflict(client, ctx, err, namespace)
	}
	if strings.Contains(errStr, "timeout") {
		return s.handleTimeout(client, ctx, err, namespace)
	}
	if strings.Contains(errStr, "failed to install") {
		return s.handleInstallFailure(client, ctx, err, namespace)
	}
	return err
}

func (s *Service) handleHelmConflict(client client.Client, ctx context.Context, originalErr error, namespace string) error {
	// try force cleanup in specified namespace
	if err := s.forceCleanupRelease(client, ctx, "velero", namespace); err != nil {
		return s.buildInstallationError("helm_conflict",
			"Helm release name conflict",
			originalErr.Error(),
			[]string{"Try running with force=true", "Manually delete existing release"},
			[]string{fmt.Sprintf("helm uninstall velero -n %s", namespace), fmt.Sprintf("kubectl delete namespace %s", namespace)})
	}
	return s.installVeleroViaHelm(client, ctx, namespace)
}

func (s *Service) handleTimeout(client client.Client, ctx context.Context, originalErr error, namespace string) error {
	extendedCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return s.installVeleroViaHelm(client, extendedCtx, namespace)
}

func (s *Service) handleInstallFailure(client client.Client, ctx context.Context, originalErr error, namespace string) error {
	if err := s.safeCleanupRelease(client, ctx, "velero"); err != nil {
		return s.buildInstallationError("install_failure",
			"Velero installation failed",
			originalErr.Error(),
			[]string{"Check cluster resources", "Verify Helm chart compatibility"},
			[]string{fmt.Sprintf("kubectl get pods -n %s", namespace), "helm list"})
	}
	return s.installVeleroViaHelm(client, ctx, namespace)
}

func (s *Service) buildInstallationError(errorType, message, details string, suggestions, commands []string) error {
	return &InstallationError{
		Type:        errorType,
		Message:     message,
		Details:     details,
		Suggestions: suggestions,
		Commands:    commands,
	}
}

//
// ─── Namespace / Resource 검사 및 정리 ──────────────────────────────────────────
//

func (s *Service) createNamespaceWithForce(client client.Client, ctx context.Context, namespace string, force bool) error {
	existingNs, err := client.Kubernetes().GetNamespace(ctx, namespace)
	if err == nil {
		return s.handleExistingNamespaceWithForce(client, ctx, namespace, existingNs, force)
	}
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err = client.Kubernetes().CreateNamespace(ctx, ns)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create namespace '%s': %w", namespace, err)
	}
	return nil
}

func (s *Service) handleExistingNamespaceWithForce(client client.Client, ctx context.Context, namespace string, existingNs *v1.Namespace, force bool) error {
	if existingNs.Status.Phase != v1.NamespaceActive {
		return fmt.Errorf("namespace '%s' is not active (phase: %s)", namespace, existingNs.Status.Phase)
	}
	veleroResources, err := s.checkVeleroResourcesInNamespace(client, ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to check existing resources in namespace '%s': %w", namespace, err)
	}
	if veleroResources.HasResources {
		if force {
			// 안전을 위해 실제 삭제 로직은 사용자 확인 필요 -> 여기서는 에러로 남김
			return s.cleanupExistingResources(client, ctx, namespace, veleroResources)
		}
		return s.handleExistingVeleroResources(client, ctx, namespace, veleroResources)
	}
	return nil
}

func (s *Service) checkVeleroResourcesInNamespace(client client.Client, ctx context.Context, namespace string) (*VeleroResources, error) {
	resources := &VeleroResources{ResourceTypes: []string{}}

	// Pods
	if rawPods, err := client.Kubernetes().GetPods(ctx, namespace, ""); err == nil {
		if podList, ok := rawPods.(*v1.PodList); ok && len(podList.Items) > 0 {
			resources.HasPods = true
			resources.ResourceCount += len(podList.Items)
			resources.ResourceTypes = append(resources.ResourceTypes, "pods")
		}
	}
	// Secrets
	if rawSecrets, err := client.Kubernetes().GetSecrets(ctx, namespace, ""); err == nil {
		if secretsList, ok := rawSecrets.([]v1.Secret); ok && len(secretsList) > 0 {
			resources.HasSecrets = true
			resources.ResourceCount += len(secretsList)
			resources.ResourceTypes = append(resources.ResourceTypes, "secrets")
		}
	}
	// ConfigMaps
	if rawCM, err := client.Kubernetes().GetConfigMaps(ctx, namespace, ""); err == nil {
		if cmList, ok := rawCM.([]v1.ConfigMap); ok && len(cmList) > 0 {
			resources.HasConfigMaps = true
			resources.ResourceCount += len(cmList)
			resources.ResourceTypes = append(resources.ResourceTypes, "configmaps")
		}
	}
	if resources.HasPods {
		resources.HasDeployments = true
		resources.ResourceTypes = append(resources.ResourceTypes, "deployments")
	}
	resources.HasResources = resources.ResourceCount > 0
	return resources, nil
}

func (s *Service) handleExistingVeleroResources(client client.Client, ctx context.Context, namespace string, resources *VeleroResources) error {
	veleroInstalled, err := s.checkVeleroInstallation(client, ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to check velero installation: %w", err)
	}
	if veleroInstalled {
		return fmt.Errorf("velero is already installed in namespace '%s' with %d resources (%v)",
			namespace, resources.ResourceCount, resources.ResourceTypes)
	}
	return fmt.Errorf("namespace '%s' contains %d existing resources (%v) but velero is not installed. Please clean up the namespace first",
		namespace, resources.ResourceCount, resources.ResourceTypes)
}

func (s *Service) cleanupExistingResources(client client.Client, ctx context.Context, namespace string, resources *VeleroResources) error {
	veleroInstalled, err := s.checkVeleroInstallation(client, ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to check velero installation: %w", err)
	}
	if veleroInstalled {
		if err := s.enhancedCleanupRelease(client, ctx, "velero"); err != nil {
			return fmt.Errorf("failed to cleanup existing velero installation: %w", err)
		}
	}
	// 안전을 위해 실제 네임스페이스 리소스 자동 삭제는 수행하지 않고 사용자에게 안내
	return fmt.Errorf("namespace '%s' contains %d existing resources (%v). Please manually clean up the namespace or use a different namespace",
		namespace, resources.ResourceCount, resources.ResourceTypes)
}

func (s *Service) checkVeleroInstallation(client client.Client, ctx context.Context, namespace string) (bool, error) {
	st, err := s.checkVeleroStatus(client, ctx, namespace)
	if err != nil {
		return false, err
	}
	return st.IsHealthy, nil
}

//
// ─── Velero readiness wait ────────────────────────────────────────────────────────
//

func (s *Service) waitForVeleroReady(client client.Client, ctx context.Context, namespace string) error {
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
			raw, err := client.Kubernetes().GetPods(ctx, namespace, "")
			if err != nil {
				continue
			}
			allReady := true
			veleroPodsFound := false
			if podList, ok := raw.(*v1.PodList); ok {
				for _, pod := range podList.Items {
					if pod.Labels["component"] == "velero" && pod.Labels["deploy"] == "velero" {
						veleroPodsFound = true
						if pod.Status.Phase != v1.PodRunning {
							allReady = false
							break
						}
					}
				}
			}
			if veleroPodsFound && allReady {
				return nil
			}
		}
	}
}

//
// ─── MinIO Secret / BSL 관련 ────────────────────────────────────────────────────
//

func (s *Service) createMinIOSecret(client client.Client, ctx context.Context, cfg config.MinioConfig, namespace string, force bool) error {
	secretName := "cloud-credentials"

	_, err := client.Kubernetes().GetSecrets(ctx, namespace, secretName)
	if err == nil && !force {
		return nil
	}
	if err == nil && force {
		if err := client.Kubernetes().DeleteSecret(ctx, namespace, secretName); err != nil {
			return fmt.Errorf("failed to delete existing secret '%s': %w", secretName, err)
		}
	}

	secretData := map[string]string{
		"cloud": fmt.Sprintf(`[default]
aws_access_key_id=%s
aws_secret_access_key=%s
region=us-west-2
`, cfg.AccessKey, cfg.SecretKey),
	}

	if _, err := client.Kubernetes().CreateSecret(ctx, namespace, secretName, secretData); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create minio secret '%s': %w", secretName, err)
	}
	return nil
}

func (s *Service) ensureMinIOSecret(client client.Client, ctx context.Context, cfg config.MinioConfig, namespace string, force bool, result *InstallResult) error {
	if err := s.createMinIOSecret(client, ctx, cfg, namespace, force); err != nil {
		return err
	}
	if force {
		result.Details["minio_secret"] = "recreated"
	} else {
		result.Details["minio_secret"] = "created"
	}
	return nil
}

func (s *Service) checkAndCreateBackupStorageLocation(client client.Client, ctx context.Context, minioCfg config.MinioConfig, namespace string, force bool) error {
	bslName := "minio"
	bsl, err := client.Velero().GetBackupStorageLocation(ctx, namespace, bslName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return s.createBackupStorageLocation(client, ctx, minioCfg, namespace, force)
		}
		return fmt.Errorf("failed to get BackupStorageLocation '%s': %w", bslName, err)
	}

	if bsl.Status.Phase != velerov1.BackupStorageLocationPhaseAvailable {
		if force {
			return s.createBackupStorageLocation(client, ctx, minioCfg, namespace, force)
		}
		statusInfo := s.buildBSLStatusInfo(bsl, minioCfg)
		errorMsg := s.buildBSLErrorMessage(bslName, string(bsl.Status.Phase), statusInfo)
		return fmt.Errorf("%s", errorMsg)
	}
	return nil
}

func (s *Service) ensureBackupStorageLocation(client client.Client, ctx context.Context, cfg config.MinioConfig, namespace string, force bool, result *InstallResult) error {
	if err := s.checkAndCreateBackupStorageLocation(client, ctx, cfg, namespace, force); err != nil {
		return err
	}
	if force {
		result.Details["backup_location"] = "recreated"
	} else {
		result.Details["backup_location"] = "created"
	}
	return nil
}

func (s *Service) createBackupStorageLocation(client client.Client, ctx context.Context, minioCfg config.MinioConfig, namespace string, force bool) error {
	bslName := "minio"

	_, err := client.Velero().GetBackupStorageLocation(ctx, namespace, bslName)
	if err == nil && !force {
		return nil
	}
	if err == nil && force {
		if err := client.Velero().DeleteBackupStorageLocation(ctx, namespace, bslName); err != nil {
			return fmt.Errorf("failed to delete existing BackupStorageLocation '%s': %w", bslName, err)
		}
	}

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
				"s3Url":            fmt.Sprintf("http://%s", minioCfg.Endpoint),
				"s3ForcePathStyle": "true",
			},
		},
	}

	if err := client.Velero().CreateBackupStorageLocation(ctx, namespace, bsl); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create BackupStorageLocation '%s': %w", bslName, err)
	}
	return nil
}

func (s *Service) validateMinIOConnection(client client.Client, ctx context.Context, minioCfg config.MinioConfig, namespace string) (bool, error) {
	if _, err := client.Minio().ListBuckets(ctx); err != nil {
		return false, fmt.Errorf("minio connection failed: %w. Please check MinIO endpoint and credentials", err)
	}
	bsl, err := client.Velero().GetBackupStorageLocation(ctx, namespace, "minio")
	if err != nil {
		return false, fmt.Errorf("failed to get BackupStorageLocation 'minio': %w", err)
	}
	if bsl.Status.Phase == velerov1.BackupStorageLocationPhaseAvailable {
		return true, nil
	}
	return false, fmt.Errorf("backup storage location 'minio' is not available. Current phase: %s. Please check MinIO connection and credentials", bsl.Status.Phase)
}

//
// ─── BSL 오류 메시지 분석 유틸 ───────────────────────────────────────────────────
//

func (s *Service) buildBSLStatusInfo(bsl *velerov1.BackupStorageLocation, minioCfg config.MinioConfig) BSLStatusInfo {
	info := BSLStatusInfo{
		Phase:            string(bsl.Status.Phase),
		Message:          bsl.Status.Message,
		MinioEndpoint:    minioCfg.Endpoint,
		Bucket:           "velero",
		Region:           "us-west-2",
		SuggestedActions: s.suggestBSLActions(bsl.Status.Message, minioCfg),
	}
	if !bsl.Status.LastValidationTime.IsZero() {
		info.LastValidationTime = bsl.Status.LastValidationTime.Format(time.RFC3339)
	}
	return info
}

func (s *Service) suggestBSLActions(message string, minioCfg config.MinioConfig) []string {
	actions := []string{}
	if strings.Contains(message, "no Host in request URL") || strings.Contains(message, "connection refused") {
		actions = append(actions, "Check MinIO endpoint URL format and accessibility")
		actions = append(actions, fmt.Sprintf("Verify MinIO is running at: %s", minioCfg.Endpoint))
	}
	if strings.Contains(message, "access denied") || strings.Contains(message, "invalid credentials") {
		actions = append(actions, "Verify MinIO AccessKey and SecretKey are correct")
		actions = append(actions, "Check MinIO user permissions for the 'velero' bucket")
	}
	if strings.Contains(message, "bucket") && (strings.Contains(message, "not found") || strings.Contains(message, "does not exist")) {
		actions = append(actions, "Create the 'velero' bucket in MinIO")
		actions = append(actions, "Verify bucket name is 'velero' (case-sensitive)")
	}
	if strings.Contains(message, "timeout") || strings.Contains(message, "exceeded maximum number of attempts") {
		actions = append(actions, "Check network connectivity to MinIO")
		actions = append(actions, "Verify MinIO server is responsive")
	}
	if len(actions) == 0 {
		actions = append(actions, "Check MinIO server status and connectivity", "Verify MinIO credentials and permissions", "Check Velero configuration")
	}
	return actions
}

func (s *Service) buildBSLErrorMessage(bslName, phase string, statusInfo BSLStatusInfo) string {
	coreError := s.extractCoreError(statusInfo.Message)
	topActions := s.selectTopActions(statusInfo.SuggestedActions)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("BackupStorageLocation '%s' is %s", bslName, strings.ToLower(phase)))
	if coreError != "" {
		b.WriteString(fmt.Sprintf(": %s", coreError))
	}
	if len(topActions) > 0 {
		b.WriteString(fmt.Sprintf(" | Try: %s", strings.Join(topActions, ", ")))
	}
	return b.String()
}

func (s *Service) extractCoreError(message string) string {
	if message == "" {
		return ""
	}
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
	parts := strings.Split(message, ":")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return message
}

func (s *Service) selectTopActions(actions []string) []string {
	if len(actions) <= 3 {
		return actions
	}
	priority := []string{}
	for _, a := range actions {
		if strings.Contains(a, "endpoint") || strings.Contains(a, "running") {
			priority = append(priority, a)
		}
	}
	if len(priority) < 3 {
		for _, a := range actions {
			if !contains(priority, a) {
				priority = append(priority, a)
				if len(priority) >= 3 {
					break
				}
			}
		}
	}
	if len(priority) > 3 {
		return priority[:3]
	}
	return priority
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetBackupsInternal : Velero Backup 목록 조회
func (s *Service) GetBackupsInternal(client client.Client, ctx context.Context, namespace string) ([]velerov1.Backup, error) {
	backups, err := client.Velero().GetBackups(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}
	return backups, nil
}

// GetRestoresInternal : Velero Restore 목록 조회
func (s *Service) GetRestoresInternal(client client.Client, ctx context.Context, namespace string) ([]velerov1.Restore, error) {
	restores, err := client.Velero().GetRestores(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list restores: %w", err)
	}
	return restores, nil
}

// GetBackupRepositoriesInternal : Velero BackupRepository 목록 조회
func (s *Service) GetBackupRepositoriesInternal(client client.Client, ctx context.Context, namespace string) ([]velerov1.BackupRepository, error) {
	repos, err := client.Velero().GetBackupRepositories(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup repositories: %w", err)
	}
	return repos, nil
}

// GetBackupStorageLocationsInternal : Velero BSL 목록 조회
func (s *Service) GetBackupStorageLocationsInternal(client client.Client, ctx context.Context, namespace string) ([]velerov1.BackupStorageLocation, error) {
	bsls, err := client.Velero().GetBackupStorageLocations(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup storage locations: %w", err)
	}
	return bsls, nil
}

// GetVolumeSnapshotLocationsInternal : Velero VSL 목록 조회
func (s *Service) GetVolumeSnapshotLocationsInternal(client client.Client, ctx context.Context, namespace string) ([]velerov1.VolumeSnapshotLocation, error) {
	vsls, err := client.Velero().GetVolumeSnapshotLocations(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list volume snapshot locations: %w", err)
	}
	return vsls, nil
}

// GetPodVolumeRestoresInternal : Velero PodVolumeRestore 목록 조회
func (s *Service) GetPodVolumeRestoresInternal(client client.Client, ctx context.Context, namespace string) ([]velerov1.PodVolumeRestore, error) {
	pvrs, err := client.Velero().GetPodVolumeRestores(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list pod volume restores: %w", err)
	}
	return pvrs, nil
}
