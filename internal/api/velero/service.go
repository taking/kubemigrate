package velero

import (
	"context"
	"fmt"
	"time"

	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/installer"
	"github.com/taking/kubemigrate/internal/job"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

// Service : Velero 관련 비즈니스 로직
type Service struct {
	*handler.BaseHandler
	jobManager job.JobManager
	installer  installer.InstallerService
}

// NewService : 새로운 Velero 서비스 생성
func NewService(base *handler.BaseHandler) *Service {
	// ConfigManager를 활용하여 워커 수 설정
	workerCount := base.GetConfigInt("VELERO_WORKER_COUNT", 3)

	return &Service{
		BaseHandler: base,
		jobManager:  job.NewMemoryJobManagerWithWorkers(workerCount),
		installer:   installer.NewService(),
	}
}

// InstallVeleroWithMinIOInternal : Velero 설치 및 MinIO 연동 설정 (비동기)
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
	go s.installVeleroWithMinIOInternal(client, ctx, jobID, cfg, namespace, force)

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

// installVeleroWithMinIOInternal : 백그라운드에서 Velero 설치
func (s *Service) installVeleroWithMinIOInternal(
	client client.Client,
	ctx context.Context,
	jobID string,
	cfg config.VeleroConfig,
	namespace string,
	force bool,
) {
	// 백그라운드 작업을 위한 새로운 context 생성 (30분 timeout)
	// ConfigManager를 활용하여 타임아웃 설정
	timeout := s.GetConfigDuration("VELERO_INSTALL_TIMEOUT", 30*time.Minute)
	bgCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Installer 설정
	installConfig := installer.VeleroInstallConfig{
		MinioConfig: cfg.MinioConfig,
		Namespace:   namespace,
		Force:       force,
	}

	// Installer를 통한 설치 실행
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 10, "Starting Velero installation...")
	s.jobManager.AddJobLog(jobID, "Starting Velero installation process")

	result, err := s.installer.InstallVelero(bgCtx, client, installConfig)
	if err != nil {
		s.jobManager.FailJob(jobID, err)
		return
	}

	// 설치 완료
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
