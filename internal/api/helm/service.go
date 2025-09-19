// Package helm Helm 관련 비즈니스 로직을 관리합니다.
package helm

import (
	"context"
	"fmt"
	"time"

	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/job"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
	v1 "k8s.io/api/core/v1"
)

// Service : Helm 관련 비즈니스 로직 서비스
type Service struct {
	*handler.BaseHandler
	jobManager job.JobManager
}

// NewService : 새로운 Helm 서비스 생성
func NewService(base *handler.BaseHandler) *Service {
	// ConfigManager를 활용하여 워커 수 설정
	workerCount := base.GetConfigInt("HELM_WORKER_COUNT", 5)

	return &Service{
		BaseHandler: base,
		jobManager:  job.NewMemoryJobManagerWithWorkers(workerCount),
	}
}

// GetChartsInternal : Helm 차트 목록 조회 (내부 로직)
func (s *Service) GetChartsInternal(client client.Client, ctx context.Context, namespace string) (interface{}, error) {
	// Helm 차트 목록 조회
	charts, err := client.Helm().GetCharts(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return charts, nil
}

// GetChartInternal : 특정 Helm 차트 상세 조회 (내부 로직)
func (s *Service) GetChartInternal(client client.Client, ctx context.Context, chartName, namespace string, version int) (interface{}, error) {
	// 특정 차트 조회
	chart, err := client.Helm().GetChart(ctx, chartName, namespace, version)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

// GetChartStatusInternal : 차트 상태 조회 (내부 로직)
func (s *Service) GetChartStatusInternal(client client.Client, ctx context.Context, chartName, namespace string) (interface{}, error) {
	// 차트 상태 조회
	chart, err := client.Helm().GetChart(ctx, chartName, namespace, 0)
	if err != nil {
		return nil, err
	}

	// 상태 정보 구성
	status := map[string]interface{}{
		"name":        chart.Name,
		"namespace":   chart.Namespace,
		"status":      chart.Info.Status,
		"version":     chart.Version,
		"revision":    chart.Version,
		"chart":       chart.Chart.Metadata.Name,
		"app_version": chart.Chart.Metadata.AppVersion,
		"updated":     chart.Info.LastDeployed,
		"description": chart.Info.Description,
	}

	return status, nil
}

// InstallChartInternal : Helm 차트 설치 (내부 로직)
func (s *Service) InstallChartInternal(client client.Client, ctx context.Context, config config.InstallChartConfig) (interface{}, error) {
	// Helm 차트 설치
	err := client.Helm().InstallChart(config.ReleaseName, config.ChartURL, config.Version, config.Namespace, config.Values)
	if err != nil {
		return nil, err
	}

	// 설치 결과 반환
	result := map[string]interface{}{
		"release_name": config.ReleaseName,
		"chart_url":    config.ChartURL,
		"version":      config.Version,
		"namespace":    config.Namespace,
		"status":       "installed",
	}

	return result, nil
}

// UpgradeChartInternal : Helm 차트 업그레이드 (내부 로직)
func (s *Service) UpgradeChartInternal(client client.Client, ctx context.Context, config config.UpgradeChartConfig) (interface{}, error) {
	// Helm 차트 업그레이드
	err := client.Helm().UpgradeChart(config.ReleaseName, config.ChartPath, "", config.Namespace, config.Values)
	if err != nil {
		return nil, err
	}

	// 업그레이드 결과 반환
	result := map[string]interface{}{
		"release_name": config.ReleaseName,
		"chart_path":   config.ChartPath,
		"namespace":    config.Namespace,
		"status":       "upgraded",
	}

	return result, nil
}

// UninstallChartInternal : Helm 차트 제거 (내부 로직)
func (s *Service) UninstallChartInternal(client client.Client, ctx context.Context, releaseName, namespace string, dryRun bool) (interface{}, error) {
	// Helm 차트 제거
	err := client.Helm().UninstallChart(releaseName, namespace, dryRun)
	if err != nil {
		return nil, err
	}

	// 제거 결과 반환
	result := map[string]interface{}{
		"release_name": releaseName,
		"namespace":    namespace,
		"dry_run":      dryRun,
		"status":       "uninstalled",
	}

	return result, nil
}

// GetChartHistoryInternal : 차트 히스토리 조회 (내부 로직)
func (s *Service) GetChartHistoryInternal(client client.Client, ctx context.Context, chartName, namespace string) (interface{}, error) {
	// 차트 히스토리 조회 (최대 10개 버전)
	history := make([]interface{}, 0)
	for i := 1; i <= 10; i++ {
		chart, err := client.Helm().GetChart(ctx, chartName, namespace, i)
		if err != nil {
			// 더 이상 히스토리가 없으면 중단
			break
		}
		history = append(history, chart)
	}

	return map[string]interface{}{
		"chart_name": chartName,
		"namespace":  namespace,
		"history":    history,
		"count":      len(history),
	}, nil
}

// GetChartValuesInternal : 차트 값 조회 (내부 로직)
func (s *Service) GetChartValuesInternal(client client.Client, ctx context.Context, chartName, namespace string) (interface{}, error) {
	// 차트 조회 (최신 버전)
	chart, err := client.Helm().GetChart(ctx, chartName, namespace, 0)
	if err != nil {
		return nil, err
	}

	// Helm의 GetValues 액션을 사용하여 실제 values 조회
	values, err := client.Helm().GetValues(ctx, chartName, namespace)
	if err != nil {
		// GetValues 실패 시 차트의 Config 사용
		if chart.Config != nil {
			values = chart.Config
		} else {
			values = map[string]interface{}{}
		}
	}

	result := map[string]interface{}{
		"name":        chartName,
		"namespace":   namespace,
		"values":      values,
		"version":     chart.Version,
		"status":      chart.Info.Status,
		"chart_name":  chart.Chart.Metadata.Name,
		"app_version": chart.Chart.Metadata.AppVersion,
	}

	return result, nil
}

// InstallChartAsyncInternal : Helm 차트 비동기 설치 (내부 로직)
func (s *Service) InstallChartAsyncInternal(client client.Client, ctx context.Context, config config.InstallChartConfig) (interface{}, error) {
	// Job ID 생성
	jobID := fmt.Sprintf("install-%d", time.Now().UnixNano())

	// Job 생성
	metadata := map[string]interface{}{
		"releaseName": config.ReleaseName,
		"chartURL":    config.ChartURL,
		"version":     config.Version,
		"namespace":   config.Namespace,
		"values":      config.Values,
	}

	job := s.jobManager.CreateJob(jobID, metadata)

	// 백그라운드에서 설치 시작
	go s.installChartInternal(client, ctx, jobID, config)

	// 즉시 응답 반환
	return map[string]interface{}{
		"status":    "processing",
		"jobId":     jobID,
		"message":   "Chart installation started",
		"statusUrl": fmt.Sprintf("/api/v1/helm/charts/status/%s", jobID),
		"logsUrl":   fmt.Sprintf("/api/v1/helm/charts/logs/%s", jobID),
		"job":       job,
	}, nil
}

// installChartInternal : 백그라운드에서 차트 설치
func (s *Service) installChartInternal(client client.Client, ctx context.Context, jobID string, config config.InstallChartConfig) {
	// 1. 다운로드 단계
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 10, "Downloading chart...")
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Starting download of chart: %s", config.ChartURL))

	// 2. 설치 단계
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 50, "Installing chart...")
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Installing chart %s in namespace %s", config.ReleaseName, config.Namespace))

	// 3. 실제 설치 실행 (재시도 로직 포함)
	err := s.jobManager.RetryOperation(jobID, "Chart installation", 3, func() error {
		return client.Helm().InstallChart(config.ReleaseName, config.ChartURL, config.Version, config.Namespace, config.Values)
	})

	if err != nil {
		s.jobManager.FailJob(jobID, err)
		return
	}

	// 4. 완료 확인 (백그라운드에서 폴링)
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 70, "Verifying installation...")
	s.jobManager.AddJobLog(jobID, "Install command executed, verifying completion...")

	// 5. 완료 대기 (최대 5분)
	if err := s.waitForInstallComplete(client, config.ReleaseName, config.Namespace, jobID, 5*time.Minute); err != nil {
		s.jobManager.FailJob(jobID, err)
		s.jobManager.AddJobLog(jobID, fmt.Sprintf("Install verification failed: %s", err.Error()))
		return
	}

	// 6. 최종 완료
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 90, "Installation completed successfully")
	s.jobManager.AddJobLog(jobID, "Chart installed successfully")

	// 7. 결과 저장
	result := map[string]interface{}{
		"releaseName": config.ReleaseName,
		"namespace":   config.Namespace,
		"version":     config.Version,
		"status":      "deployed",
		"installedAt": time.Now().Format(time.RFC3339),
	}

	s.jobManager.CompleteJob(jobID, result)
	s.jobManager.AddJobLog(jobID, "Installation completed successfully")
}

// UpgradeChartAsyncInternal : Helm 차트 비동기 업그레이드 (내부 로직)
func (s *Service) UpgradeChartAsyncInternal(client client.Client, ctx context.Context, config config.UpgradeChartConfig) (interface{}, error) {
	// Job ID 생성
	jobID := fmt.Sprintf("upgrade-%d", time.Now().UnixNano())

	// Job 생성
	metadata := map[string]interface{}{
		"releaseName": config.ReleaseName,
		"chartPath":   config.ChartPath,
		"namespace":   config.Namespace,
		"values":      config.Values,
	}

	job := s.jobManager.CreateJob(jobID, metadata)

	// 백그라운드에서 업그레이드 시작
	go s.upgradeChartInternal(client, ctx, jobID, config)

	// 즉시 응답 반환
	return map[string]interface{}{
		"status":    "processing",
		"jobId":     jobID,
		"message":   "Chart upgrade started",
		"statusUrl": fmt.Sprintf("/api/v1/helm/charts/status/%s", jobID),
		"logsUrl":   fmt.Sprintf("/api/v1/helm/charts/logs/%s", jobID),
		"job":       job,
	}, nil
}

// upgradeChartInternal : 백그라운드에서 차트 업그레이드
func (s *Service) upgradeChartInternal(client client.Client, ctx context.Context, jobID string, config config.UpgradeChartConfig) {
	// 1. 업그레이드 준비 단계
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 10, "Preparing upgrade...")
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Starting upgrade of chart: %s", config.ReleaseName))

	// 2. 업그레이드 실행 단계
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 50, "Upgrading chart...")
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Upgrading chart %s in namespace %s", config.ReleaseName, config.Namespace))

	// 3. 실제 업그레이드 실행 (재시도 로직 포함)
	err := s.jobManager.RetryOperation(jobID, "Chart upgrade", 3, func() error {
		return client.Helm().UpgradeChart(config.ReleaseName, config.ChartPath, "", config.Namespace, config.Values)
	})

	if err != nil {
		s.jobManager.FailJob(jobID, err)
		return
	}

	// 4. 완료 확인 (백그라운드에서 폴링)
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 70, "Verifying upgrade...")
	s.jobManager.AddJobLog(jobID, "Upgrade command executed, verifying completion...")

	// 5. 완료 대기 (최대 5분)
	if err := s.waitForUpgradeComplete(client, config.ReleaseName, config.Namespace, jobID, 5*time.Minute); err != nil {
		s.jobManager.FailJob(jobID, err)
		s.jobManager.AddJobLog(jobID, fmt.Sprintf("Upgrade verification failed: %s", err.Error()))
		return
	}

	// 6. 최종 완료
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 90, "Upgrade completed successfully")
	s.jobManager.AddJobLog(jobID, "Chart upgraded successfully")

	// 7. 결과 저장
	result := map[string]interface{}{
		"releaseName": config.ReleaseName,
		"namespace":   config.Namespace,
		"status":      "upgraded",
		"upgradedAt":  time.Now().Format(time.RFC3339),
	}

	s.jobManager.CompleteJob(jobID, result)
	s.jobManager.AddJobLog(jobID, "Upgrade completed successfully")
}

// UninstallChartAsyncInternal : Helm 차트 비동기 제거 (내부 로직)
func (s *Service) UninstallChartAsyncInternal(client client.Client, ctx context.Context, releaseName, namespace string, dryRun bool) (interface{}, error) {
	// Job ID 생성
	jobID := fmt.Sprintf("uninstall-%d", time.Now().UnixNano())

	// Job 생성
	metadata := map[string]interface{}{
		"releaseName": releaseName,
		"namespace":   namespace,
		"dryRun":      dryRun,
	}

	job := s.jobManager.CreateJob(jobID, metadata)

	// 백그라운드에서 제거 시작
	go s.uninstallChartInternal(client, ctx, jobID, releaseName, namespace, dryRun)

	// 즉시 응답 반환
	return map[string]interface{}{
		"status":    "processing",
		"jobId":     jobID,
		"message":   "Chart uninstall started",
		"statusUrl": fmt.Sprintf("/api/v1/helm/charts/status/%s", jobID),
		"logsUrl":   fmt.Sprintf("/api/v1/helm/charts/logs/%s", jobID),
		"job":       job,
	}, nil
}

// uninstallChartInternal : 백그라운드에서 차트 제거
func (s *Service) uninstallChartInternal(client client.Client, ctx context.Context, jobID, releaseName, namespace string, dryRun bool) {
	// 1. 제거 준비 단계
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 10, "Preparing uninstall...")
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Starting uninstall of chart: %s", releaseName))

	// 2. 제거 실행 단계
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 50, "Uninstalling chart...")
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Uninstalling chart %s from namespace %s (dryRun: %v)", releaseName, namespace, dryRun))

	// 3. 실제 제거 실행 (재시도 로직 포함)
	err := s.jobManager.RetryOperation(jobID, "Chart uninstall", 3, func() error {
		return client.Helm().UninstallChart(releaseName, namespace, dryRun)
	})

	if err != nil {
		s.jobManager.FailJob(jobID, err)
		return
	}

	// 4. 완료 확인 (백그라운드에서 폴링)
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 70, "Verifying uninstall...")
	s.jobManager.AddJobLog(jobID, "Uninstall command executed, verifying completion...")

	// 5. 완료 대기 (최대 5분)
	if err := s.waitForUninstallComplete(client, releaseName, namespace, jobID, 5*time.Minute); err != nil {
		s.jobManager.FailJob(jobID, err)
		s.jobManager.AddJobLog(jobID, fmt.Sprintf("Uninstall verification failed: %s", err.Error()))
		return
	}

	// 6. 최종 완료
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 90, "Uninstall completed successfully")
	s.jobManager.AddJobLog(jobID, "Chart uninstalled successfully")

	// 5. 결과 저장
	result := map[string]interface{}{
		"releaseName":   releaseName,
		"namespace":     namespace,
		"dryRun":        dryRun,
		"status":        "uninstalled",
		"uninstalledAt": time.Now().Format(time.RFC3339),
	}

	s.jobManager.CompleteJob(jobID, result)
	s.jobManager.AddJobLog(jobID, "Uninstall completed successfully")
}

// waitForUninstallComplete : Uninstall 완료 대기 (Kubernetes 리소스 기반 확인)
func (s *Service) waitForUninstallComplete(client client.Client, releaseName, namespace, jobID string, timeout time.Duration) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(timeout)

	for { //nolint:gosimple
		select {
		case <-ticker.C:
			// 1. Helm Release 확인
			installed, _, err := client.Helm().IsChartInstalled(releaseName)
			if err != nil {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Error checking chart status: %s", err.Error()))
				continue
			}
			if !installed {
				s.jobManager.AddJobLog(jobID, "Chart uninstalled successfully")
				return nil
			}

			// 2. Kubernetes 리소스 확인 (더 정확한 확인)
			removed, err := s.isReleaseResourcesRemoved(client, releaseName, namespace)
			if err != nil {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Error checking Kubernetes resources: %s", err.Error()))
				continue
			}
			if removed {
				s.jobManager.AddJobLog(jobID, "All Kubernetes resources removed successfully")
				return nil
			}

			// 3. 타임아웃 확인
			if time.Now().After(deadline) {
				return fmt.Errorf("uninstall timeout after %v", timeout)
			}

			s.jobManager.AddJobLog(jobID, fmt.Sprintf("Uninstall in progress... (remaining: %v)", time.Until(deadline).Round(time.Second)))
		}
	}
}

// isReleaseResourcesRemoved : Kubernetes 리소스 기반으로 Release 제거 확인
func (s *Service) isReleaseResourcesRemoved(client client.Client, releaseName, namespace string) (bool, error) {
	ctx := context.Background()

	// 1. ConfigMap 확인 (Helm Release 정보)
	configMaps, err := client.Kubernetes().GetConfigMaps(ctx, namespace, "")
	if err != nil {
		return false, err
	}

	if cmList, ok := configMaps.([]v1.ConfigMap); ok {
		for _, configMap := range cmList {
			if configMap.Labels["NAME"] == releaseName {
				return false, nil // 아직 존재
			}
		}
	}

	// 2. Secret 확인 (Helm Release 정보)
	secrets, err := client.Kubernetes().GetSecrets(ctx, namespace, "")
	if err != nil {
		return false, err
	}

	if secretList, ok := secrets.([]v1.Secret); ok {
		for _, sec := range secretList {
			if sec.Labels["NAME"] == releaseName {
				return false, nil // 아직 존재
			}
		}
	}

	return true, nil // 모든 리소스 제거됨
}

// waitForInstallComplete : Install 완료 대기 (Kubernetes 리소스 기반 확인)
func (s *Service) waitForInstallComplete(client client.Client, releaseName, namespace, jobID string, timeout time.Duration) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(timeout)

	for { //nolint:gosimple
		select {
		case <-ticker.C:
			// 1. Helm Release 확인
			installed, _, err := client.Helm().IsChartInstalled(releaseName)
			if err != nil {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Error checking chart status: %s", err.Error()))
				continue
			}
			if installed {
				s.jobManager.AddJobLog(jobID, "Chart installed successfully")
				return nil
			}

			// 2. Kubernetes 리소스 확인 (더 정확한 확인)
			installed, err = s.isReleaseResourcesInstalled(client, releaseName, namespace)
			if err != nil {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Error checking Kubernetes resources: %s", err.Error()))
				continue
			}
			if installed {
				s.jobManager.AddJobLog(jobID, "All Kubernetes resources installed successfully")
				return nil
			}

			// 3. 타임아웃 확인
			if time.Now().After(deadline) {
				return fmt.Errorf("install timeout after %v", timeout)
			}

			s.jobManager.AddJobLog(jobID, fmt.Sprintf("Install in progress... (remaining: %v)", time.Until(deadline).Round(time.Second)))
		}
	}
}

// waitForUpgradeComplete : Upgrade 완료 대기 (Kubernetes 리소스 기반 확인)
func (s *Service) waitForUpgradeComplete(client client.Client, releaseName, namespace, jobID string, timeout time.Duration) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(timeout)

	for { //nolint:gosimple
		select {
		case <-ticker.C:
			// 1. Helm Release 확인
			installed, _, err := client.Helm().IsChartInstalled(releaseName)
			if err != nil {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Error checking chart status: %s", err.Error()))
				continue
			}
			if installed {
				s.jobManager.AddJobLog(jobID, "Chart upgraded successfully")
				return nil
			}

			// 2. Kubernetes 리소스 확인 (더 정확한 확인)
			installed, err = s.isReleaseResourcesInstalled(client, releaseName, namespace)
			if err != nil {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Error checking Kubernetes resources: %s", err.Error()))
				continue
			}
			if installed {
				s.jobManager.AddJobLog(jobID, "All Kubernetes resources upgraded successfully")
				return nil
			}

			// 3. 타임아웃 확인
			if time.Now().After(deadline) {
				return fmt.Errorf("upgrade timeout after %v", timeout)
			}

			s.jobManager.AddJobLog(jobID, fmt.Sprintf("Upgrade in progress... (remaining: %v)", time.Until(deadline).Round(time.Second)))
		}
	}
}

// isReleaseResourcesInstalled : Kubernetes 리소스 기반으로 Release 설치 확인
func (s *Service) isReleaseResourcesInstalled(client client.Client, releaseName, namespace string) (bool, error) {
	ctx := context.Background()

	// 1. ConfigMap 확인 (Helm Release 정보)
	configMaps, err := client.Kubernetes().GetConfigMaps(ctx, namespace, "")
	if err != nil {
		return false, err
	}

	if cmList, ok := configMaps.([]v1.ConfigMap); ok {
		for _, configMap := range cmList {
			if configMap.Labels["NAME"] == releaseName {
				return true, nil // 설치됨
			}
		}
	}

	// 2. Secret 확인 (Helm Release 정보)
	secrets, err := client.Kubernetes().GetSecrets(ctx, namespace, "")
	if err != nil {
		return false, err
	}

	if secretList, ok := secrets.([]v1.Secret); ok {
		for _, sec := range secretList {
			if sec.Labels["NAME"] == releaseName {
				return true, nil // 설치됨
			}
		}
	}

	return false, nil // 아직 설치되지 않음
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
