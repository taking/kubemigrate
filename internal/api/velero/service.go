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
	"github.com/taking/kubemigrate/pkg/types"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// managedFields 제거
	removeManagedFieldsFromBackups(backups)

	return backups, nil
}

// GetRestoresInternal : Velero Restore 목록 조회
func (s *Service) GetRestoresInternal(client client.Client, ctx context.Context, namespace string) ([]velerov1.Restore, error) {
	restores, err := client.Velero().GetRestores(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list restores: %w", err)
	}

	// managedFields 제거
	removeManagedFieldsFromRestores(restores)

	return restores, nil
}

// GetBackupRepositoriesInternal : Velero BackupRepository 목록 조회
func (s *Service) GetBackupRepositoriesInternal(client client.Client, ctx context.Context, namespace string) ([]velerov1.BackupRepository, error) {
	repos, err := client.Velero().GetBackupRepositories(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup repositories: %w", err)
	}

	// managedFields 제거
	removeManagedFieldsFromBackupRepositories(repos)

	return repos, nil
}

// GetBackupStorageLocationsInternal : Velero BSL 목록 조회
func (s *Service) GetBackupStorageLocationsInternal(client client.Client, ctx context.Context, namespace string) ([]velerov1.BackupStorageLocation, error) {
	bsls, err := client.Velero().GetBackupStorageLocations(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup storage locations: %w", err)
	}

	// managedFields 제거
	removeManagedFieldsFromBackupStorageLocations(bsls)

	return bsls, nil
}

// GetVolumeSnapshotLocationsInternal : Velero VSL 목록 조회
func (s *Service) GetVolumeSnapshotLocationsInternal(client client.Client, ctx context.Context, namespace string) ([]velerov1.VolumeSnapshotLocation, error) {
	vsls, err := client.Velero().GetVolumeSnapshotLocations(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list volume snapshot locations: %w", err)
	}

	// managedFields 제거
	removeManagedFieldsFromVolumeSnapshotLocations(vsls)

	return vsls, nil
}

// GetPodVolumeRestoresInternal : Velero PodVolumeRestore 목록 조회
func (s *Service) GetPodVolumeRestoresInternal(client client.Client, ctx context.Context, namespace string) ([]velerov1.PodVolumeRestore, error) {
	pvrs, err := client.Velero().GetPodVolumeRestores(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list pod volume restores: %w", err)
	}

	// managedFields 제거
	removeManagedFieldsFromPodVolumeRestores(pvrs)

	return pvrs, nil
}

// CreateBackupInternal : Velero 백업 생성 (비동기)
func (s *Service) CreateBackupInternal(
	client client.Client,
	ctx context.Context,
	backupReq types.BackupRequest,
	namespace string,
) (interface{}, error) {
	// Job ID 생성
	jobID := fmt.Sprintf("backup-create-%d", time.Now().UnixNano())

	// Job 생성 (민감한 정보 제외)
	metadata := map[string]interface{}{
		"backupName":        backupReq.Name,
		"namespace":         namespace,
		"includeNamespaces": backupReq.IncludeNamespaces,
		"excludeNamespaces": backupReq.ExcludeNamespaces,
		"includeResources":  backupReq.IncludeResources,
		"excludeResources":  backupReq.ExcludeResources,
		"storageLocation":   backupReq.StorageLocation,
		"ttl":               backupReq.TTL,
	}

	_ = s.jobManager.CreateJob(jobID, metadata)

	// 백그라운드에서 백업 생성 시작
	go s.createBackupInternal(client, ctx, jobID, backupReq, namespace)

	// 즉시 응답 반환
	return types.BackupResult{
		Status:     "processing",
		JobID:      jobID,
		BackupName: backupReq.Name,
		Namespace:  namespace,
		Message:    "Backup creation started",
		StatusUrl:  fmt.Sprintf("/api/v1/velero/status/%s", jobID),
		LogsUrl:    fmt.Sprintf("/api/v1/velero/logs/%s", jobID),
		CreatedAt:  time.Now(),
	}, nil
}

// createBackupInternal : 백그라운드에서 백업 생성
func (s *Service) createBackupInternal(
	client client.Client,
	ctx context.Context,
	jobID string,
	backupReq types.BackupRequest,
	namespace string,
) {
	// 백그라운드 작업을 위한 새로운 context 생성 (30분 timeout)
	timeout := s.GetConfigDuration("VELERO_BACKUP_TIMEOUT", 30*time.Minute)
	bgCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Velero Backup 리소스 생성
	backup := &velerov1.Backup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupReq.Name,
			Namespace: namespace,
		},
		Spec: velerov1.BackupSpec{
			IncludedNamespaces:      backupReq.IncludeNamespaces,
			ExcludedNamespaces:      backupReq.ExcludeNamespaces,
			IncludedResources:       backupReq.IncludeResources,
			ExcludedResources:       backupReq.ExcludeResources,
			LabelSelector:           &metav1.LabelSelector{MatchLabels: backupReq.LabelSelector},
			StorageLocation:         backupReq.StorageLocation,
			VolumeSnapshotLocations: backupReq.VolumeSnapshotLocations,
			TTL:                     metav1.Duration{Duration: parseTTL(backupReq.TTL)},
			IncludeClusterResources: backupReq.IncludeClusterResources,
		},
	}

	// 백업 생성 시작
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 10, "Creating Velero backup...")
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Creating backup: %s", backupReq.Name))

	// Velero 백업 리소스 생성
	err := client.Velero().CreateBackup(bgCtx, namespace, backup)
	if err != nil {
		s.jobManager.FailJob(jobID, err)
		return
	}

	// 백업 생성 완료
	result := map[string]interface{}{
		"backupName":      backupReq.Name,
		"namespace":       namespace,
		"status":          "created",
		"message":         "Backup created successfully",
		"createdAt":       time.Now(),
		"storageLocation": backupReq.StorageLocation,
		"ttl":             backupReq.TTL,
	}

	s.jobManager.CompleteJob(jobID, result)
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Backup %s created successfully", backupReq.Name))
}

// ValidateBackupInternal : 백업 검증 (동기)
func (s *Service) ValidateBackupInternal(
	client client.Client,
	ctx context.Context,
	backupName string,
	namespace string,
) (interface{}, error) {
	// 백업 조회
	backup, err := client.Velero().GetBackup(ctx, namespace, backupName)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}

	// managedFields 제거
	backup.ObjectMeta.ManagedFields = nil

	// 검증 결과 초기화
	validationResult := types.BackupValidationResult{
		BackupName:        backupName,
		ValidationTime:    time.Now(),
		Phase:             string(backup.Status.Phase),
		ValidationDetails: types.ValidationDetails{},
		Summary:           types.BackupSummary{},
	}

	// 백업 상태 검증
	isValid := true
	var errors []string
	var warnings []string

	// Phase 검증
	switch backup.Status.Phase {
	case velerov1.BackupPhaseCompleted:
		validationResult.IsValid = true
	case velerov1.BackupPhaseFailed:
		validationResult.IsValid = false
		errors = append(errors, "Backup failed")
		if backup.Status.FailureReason != "" {
			errors = append(errors, backup.Status.FailureReason)
		}
	case velerov1.BackupPhasePartiallyFailed:
		validationResult.IsValid = false
		warnings = append(warnings, "Backup partially failed")
		if backup.Status.FailureReason != "" {
			warnings = append(warnings, backup.Status.FailureReason)
		}
	case velerov1.BackupPhaseInProgress:
		validationResult.IsValid = false
		errors = append(errors, "Backup is still in progress")
	default:
		validationResult.IsValid = false
		errors = append(errors, fmt.Sprintf("Unknown backup phase: %s", backup.Status.Phase))
	}

	// 스토리지 위치 검증
	if backup.Spec.StorageLocation != "" {
		bsl, err := client.Velero().GetBackupStorageLocation(ctx, namespace, backup.Spec.StorageLocation)
		if err != nil {
			validationResult.ValidationDetails.StorageLocationValid = false
			validationResult.ValidationDetails.StorageLocationErrors = append(validationResult.ValidationDetails.StorageLocationErrors, err.Error())
			errors = append(errors, fmt.Sprintf("Storage location validation failed: %v", err))
		} else {
			validationResult.ValidationDetails.StorageLocationValid = true
			validationResult.Summary.StorageLocation = bsl.Name
		}
	}

	// 볼륨 스냅샷 위치 검증
	if len(backup.Spec.VolumeSnapshotLocations) > 0 {
		for _, vslName := range backup.Spec.VolumeSnapshotLocations {
			_, err := client.Velero().GetVolumeSnapshotLocation(ctx, namespace, vslName)
			if err != nil {
				validationResult.ValidationDetails.VolumeSnapshotValid = false
				validationResult.ValidationDetails.VolumeSnapshotErrors = append(validationResult.ValidationDetails.VolumeSnapshotErrors, err.Error())
				errors = append(errors, fmt.Sprintf("Volume snapshot location validation failed: %v", err))
			} else {
				validationResult.ValidationDetails.VolumeSnapshotValid = true
			}
		}
	}

	// 백업 요약 정보 설정 (간단한 버전)
	validationResult.Summary.TotalItems = 0
	validationResult.ValidationDetails.ResourceCount = 0
	validationResult.ValidationDetails.VolumeCount = 0

	// 최종 검증 결과 설정
	validationResult.Errors = errors
	validationResult.Warnings = warnings
	validationResult.IsValid = isValid && len(errors) == 0

	return validationResult, nil
}

// DeleteBackupInternal : Velero 백업 삭제
func (s *Service) DeleteBackupInternal(
	client client.Client,
	ctx context.Context,
	backupName string,
	namespace string,
) (interface{}, error) {
	// 백업 존재 여부 확인
	backup, err := client.Velero().GetBackup(ctx, namespace, backupName)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}

	// 백업 삭제
	err = client.Velero().DeleteBackup(ctx, namespace, backupName)
	if err != nil {
		return nil, fmt.Errorf("failed to delete backup: %w", err)
	}

	// 삭제 결과 반환
	result := map[string]interface{}{
		"backupName":    backupName,
		"namespace":     namespace,
		"status":        "deleted",
		"message":       "Backup deleted successfully",
		"deletedAt":     time.Now(),
		"backupPhase":   string(backup.Status.Phase),
		"backupCreated": backup.CreationTimestamp.Time,
	}

	return result, nil
}

// parseTTL : TTL 문자열을 time.Duration으로 변환
func parseTTL(ttl string) time.Duration {
	if ttl == "" {
		return 720 * time.Hour // 기본 30일
	}

	duration, err := time.ParseDuration(ttl)
	if err != nil {
		return 720 * time.Hour // 파싱 실패 시 기본값
	}

	return duration
}

// ===== managedFields 제거 헬퍼 함수들 =====

// removeManagedFieldsFromBackups : Backup 목록에서 managedFields 제거
func removeManagedFieldsFromBackups(backups []velerov1.Backup) {
	for i := range backups {
		backups[i].ObjectMeta.ManagedFields = nil
	}
}

// removeManagedFieldsFromRestores : Restore 목록에서 managedFields 제거
func removeManagedFieldsFromRestores(restores []velerov1.Restore) {
	for i := range restores {
		restores[i].ObjectMeta.ManagedFields = nil
	}
}

// removeManagedFieldsFromBackupRepositories : BackupRepository 목록에서 managedFields 제거
func removeManagedFieldsFromBackupRepositories(repos []velerov1.BackupRepository) {
	for i := range repos {
		repos[i].ObjectMeta.ManagedFields = nil
	}
}

// removeManagedFieldsFromBackupStorageLocations : BackupStorageLocation 목록에서 managedFields 제거
func removeManagedFieldsFromBackupStorageLocations(bsls []velerov1.BackupStorageLocation) {
	for i := range bsls {
		bsls[i].ObjectMeta.ManagedFields = nil
	}
}

// removeManagedFieldsFromVolumeSnapshotLocations : VolumeSnapshotLocation 목록에서 managedFields 제거
func removeManagedFieldsFromVolumeSnapshotLocations(vsls []velerov1.VolumeSnapshotLocation) {
	for i := range vsls {
		vsls[i].ObjectMeta.ManagedFields = nil
	}
}

// removeManagedFieldsFromPodVolumeRestores : PodVolumeRestore 목록에서 managedFields 제거
func removeManagedFieldsFromPodVolumeRestores(pvrs []velerov1.PodVolumeRestore) {
	for i := range pvrs {
		pvrs[i].ObjectMeta.ManagedFields = nil
	}
}
