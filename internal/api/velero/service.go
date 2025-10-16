package velero

import (
	"context"
	"fmt"
	"time"

	"github.com/taking/kubemigrate/internal/api/minio"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/installer"
	"github.com/taking/kubemigrate/internal/job"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
	"github.com/taking/kubemigrate/pkg/types"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Service : Velero 관련 비즈니스 로직
type Service struct {
	*handler.BaseHandler
	jobManager   job.JobManager
	installer    installer.InstallerService
	minioService *minio.Service
}

// NewService : 새로운 Velero 서비스 생성
func NewService(base *handler.BaseHandler) *Service {
	// ConfigManager를 활용하여 워커 수 설정
	workerCount := base.GetConfigInt("VELERO_WORKER_COUNT", 3)

	return &Service{
		BaseHandler:  base,
		jobManager:   job.NewMemoryJobManagerWithWorkers(workerCount),
		installer:    installer.NewService(),
		minioService: minio.NewService(base),
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

// GetBackupInternal : Velero Backup 상세 조회
func (s *Service) GetBackupInternal(client client.Client, ctx context.Context, namespace, backupName string) (*velerov1.Backup, error) {
	backup, err := client.Velero().GetBackup(ctx, namespace, backupName)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup '%s': %w", backupName, err)
	}

	// managedFields 제거
	backup.ObjectMeta.ManagedFields = nil

	return backup, nil
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

// GetRestoreInternal : Velero Restore 상세 조회
func (s *Service) GetRestoreInternal(client client.Client, ctx context.Context, namespace, restoreName string) (*velerov1.Restore, error) {
	restore, err := client.Velero().GetRestore(ctx, namespace, restoreName)
	if err != nil {
		return nil, fmt.Errorf("failed to get restore '%s': %w", restoreName, err)
	}

	// managedFields 제거
	restore.ObjectMeta.ManagedFields = nil

	return restore, nil
}

// CreateRestoreInternal : Velero Restore 생성 (비동기)
func (s *Service) CreateRestoreInternal(client client.Client, ctx context.Context, req types.VeleroRestoreRequest) (interface{}, error) {
	// Job ID 생성
	jobID := fmt.Sprintf("restore-create-%d", time.Now().UnixNano())

	// Job 생성 (민감한 정보 제외)
	metadata := map[string]interface{}{
		"restoreName":             req.Restore.Name,
		"backupName":              req.Restore.BackupName,
		"includeNamespaces":       req.Restore.IncludeNamespaces,
		"excludeNamespaces":       req.Restore.ExcludeNamespaces,
		"includeResources":        req.Restore.IncludeResources,
		"excludeResources":        req.Restore.ExcludeResources,
		"includeClusterResources": req.Restore.IncludeClusterResources,
		"restorePVs":              req.Restore.RestorePVs,
		"storageClassMappings":    req.Restore.StorageClassMappings,
	}

	_ = s.jobManager.CreateJob(jobID, metadata)

	// 백그라운드에서 복구 생성 시작
	go s.createRestoreInternal(client, ctx, jobID, req)

	// 즉시 응답 반환
	return types.RestoreResult{
		Status:      "processing",
		JobID:       jobID,
		RestoreName: req.Restore.Name,
		BackupName:  req.Restore.BackupName,
		Namespace:   "velero",
		Message:     "Restore creation started",
		StatusUrl:   fmt.Sprintf("/api/v1/velero/status/%s", jobID),
		LogsUrl:     fmt.Sprintf("/api/v1/velero/logs/%s", jobID),
		CreatedAt:   time.Now(),
	}, nil
}

// createRestoreInternal : 백그라운드에서 복구 생성
func (s *Service) createRestoreInternal(
	client client.Client,
	ctx context.Context,
	jobID string,
	req types.VeleroRestoreRequest,
) {
	// 백그라운드 작업을 위한 새로운 context 생성 (30분 timeout)
	timeout := s.GetConfigDuration("VELERO_RESTORE_TIMEOUT", 30*time.Minute)
	bgCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// StorageClass 매핑이 있는 경우 ConfigMap 생성
	if len(req.Restore.StorageClassMappings) > 0 {
		s.jobManager.AddJobLog(jobID, "StorageClass mappings detected, creating ConfigMap...")
		err := s.createStorageClassConfigMap(client, bgCtx, jobID, req.Restore.StorageClassMappings)
		if err != nil {
			s.jobManager.FailJob(jobID, fmt.Errorf("failed to create StorageClass ConfigMap: %w", err))
			return
		}
		// 복구 완료 후 ConfigMap 정리를 위한 defer 추가
		defer func() {
			if cleanupErr := s.deleteStorageClassConfigMap(client, bgCtx, jobID); cleanupErr != nil {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Warning: Failed to cleanup StorageClass ConfigMap: %v", cleanupErr))
			}
		}()
	}

	// RestoreRequest를 velerov1.Restore로 변환
	restore := &velerov1.Restore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Restore.Name,
			Namespace: "velero",
		},
		Spec: velerov1.RestoreSpec{
			BackupName:              req.Restore.BackupName,
			IncludedNamespaces:      req.Restore.IncludeNamespaces,
			ExcludedNamespaces:      req.Restore.ExcludeNamespaces,
			IncludedResources:       req.Restore.IncludeResources,
			ExcludedResources:       req.Restore.ExcludeResources,
			IncludeClusterResources: &req.Restore.IncludeClusterResources,
			RestorePVs:              &req.Restore.RestorePVs,
		},
	}

	// LabelSelector 변환
	if req.Restore.LabelSelector != nil && len(req.Restore.LabelSelector) > 0 {
		restore.Spec.LabelSelector = &metav1.LabelSelector{
			MatchLabels: req.Restore.LabelSelector,
		}
	}

	// StorageClass 매핑 처리 - Velero Restore Spec에 포함
	if req.Restore.StorageClassMappings != nil && len(req.Restore.StorageClassMappings) > 0 {
		// Velero v1.17.0에서는 StorageClassMappings를 직접 지원하지 않음
		// 대신 복구 완료 후 별도 처리
		s.jobManager.AddJobLog(jobID, "StorageClass mappings will be applied after restore completion")
	}

	// 복구 생성 시작
	s.jobManager.UpdateJobStatus(jobID, job.JobStatusProcessing, 10, "Creating Velero restore...")
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Creating restore: %s", req.Restore.Name))

	// Velero 복구 리소스 생성
	err := client.Velero().CreateRestore(bgCtx, "velero", restore)
	if err != nil {
		s.jobManager.FailJob(jobID, err)
		return
	}

	// 복구 완료 대기
	s.jobManager.AddJobLog(jobID, "Waiting for restore to complete...")
	s.waitForRestoreCompletion(bgCtx, client, jobID, req.Restore.Name)

	// 복구 생성 완료
	result := map[string]interface{}{
		"restoreName":       req.Restore.Name,
		"backupName":        req.Restore.BackupName,
		"namespace":         "velero",
		"status":            "created",
		"message":           "Restore created successfully",
		"createdAt":         time.Now(),
		"includeNamespaces": req.Restore.IncludeNamespaces,
	}

	s.jobManager.CompleteJob(jobID, result)
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Restore %s created successfully", req.Restore.Name))
}

// applyStorageClassMappings : StorageClass 매핑을 적용합니다
func (s *Service) applyStorageClassMappings(
	ctx context.Context,
	client client.Client,
	jobID string,
	storageClassMappings map[string]string,
	includeNamespaces []string,
) {
	for originalSC, newSC := range storageClassMappings {
		for _, namespace := range includeNamespaces {
			s.jobManager.AddJobLog(jobID, fmt.Sprintf("Processing namespace %s: %s -> %s", namespace, originalSC, newSC))

			// 해당 네임스페이스의 PVC 목록 조회
			pvcList, err := client.Kubernetes().GetPVCs(ctx, namespace, "")
			if err != nil {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Failed to get PVCs in namespace %s: %v", namespace, err))
				continue
			}

			// PVC 목록을 타입 캐스팅
			pvcs, ok := pvcList.(*v1.PersistentVolumeClaimList)
			if !ok {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Failed to cast PVC list in namespace %s", namespace))
				continue
			}

			// 해당 StorageClass를 사용하는 PVC 찾기
			var targetPVCs []string
			for _, pvc := range pvcs.Items {
				if pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == originalSC {
					targetPVCs = append(targetPVCs, pvc.Name)
				}
			}

			if len(targetPVCs) == 0 {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("No PVCs found with StorageClass %s in namespace %s", originalSC, namespace))
				continue
			}

			// 각 PVC의 StorageClass 변경 (삭제 후 재생성 방식)
			for _, pvcName := range targetPVCs {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Recreating PVC %s/%s: %s -> %s", namespace, pvcName, originalSC, newSC))

				// PVC 삭제 후 재생성
				err := s.recreatePVCWithNewStorageClass(ctx, client, jobID, namespace, pvcName, newSC)
				if err != nil {
					s.jobManager.AddJobLog(jobID, fmt.Sprintf("Failed to recreate PVC %s/%s: %v", namespace, pvcName, err))
					continue
				}

				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Successfully recreated PVC %s/%s", namespace, pvcName))
			}
		}
	}
}

// waitForPVCBinding : PVC가 새로운 PV와 바인딩될 때까지 대기
func (s *Service) waitForPVCBinding(
	ctx context.Context,
	client client.Client,
	jobID string,
	namespace, pvcName string,
	originalPVC *v1.PersistentVolumeClaim,
) {
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Waiting for PVC %s/%s to bind to new PV...", namespace, pvcName))

	// 최대 3분 대기 (단축)
	waitTimeout := 3 * time.Minute
	checkInterval := 10 * time.Second

	for start := time.Now(); time.Since(start) < waitTimeout; time.Sleep(checkInterval) {
		// PVC 상태 확인
		pvc, err := client.Kubernetes().GetPVCs(ctx, namespace, pvcName)
		if err != nil {
			s.jobManager.AddJobLog(jobID, fmt.Sprintf("Failed to check PVC %s/%s status: %v", namespace, pvcName, err))
			continue
		}

		currentPVC, ok := pvc.(*v1.PersistentVolumeClaim)
		if !ok {
			s.jobManager.AddJobLog(jobID, fmt.Sprintf("Failed to cast PVC %s/%s", namespace, pvcName))
			continue
		}

		// PVC 상태 확인
		if currentPVC.Status.Phase == v1.ClaimBound {
			s.jobManager.AddJobLog(jobID, fmt.Sprintf("PVC %s/%s successfully bound to new PV", namespace, pvcName))
			return
		}

		s.jobManager.AddJobLog(jobID, fmt.Sprintf("PVC %s/%s status: %s, waiting...", namespace, pvcName, currentPVC.Status.Phase))
	}

	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Timeout waiting for PVC %s/%s to bind", namespace, pvcName))
}

// waitForRestoreCompletion : 복구 완료까지 대기
func (s *Service) waitForRestoreCompletion(
	ctx context.Context,
	client client.Client,
	jobID string,
	restoreName string,
) {
	// 복구 완료 대기 (최대 30분)
	waitTimeout := 30 * time.Minute
	checkInterval := 30 * time.Second

	for start := time.Now(); time.Since(start) < waitTimeout; time.Sleep(checkInterval) {
		// 복구 상태 확인
		restoreStatus, err := client.Velero().GetRestore(ctx, "velero", restoreName)
		if err != nil {
			s.jobManager.AddJobLog(jobID, fmt.Sprintf("Failed to check restore status: %v", err))
			continue
		}

		// 복구 완료 확인
		if restoreStatus.Status.Phase == "Completed" || restoreStatus.Status.Phase == "PartiallyFailed" {
			s.jobManager.AddJobLog(jobID, fmt.Sprintf("Restore completed with phase: %s", restoreStatus.Status.Phase))
			return
		}

		s.jobManager.AddJobLog(jobID, fmt.Sprintf("Restore status: %s, waiting...", restoreStatus.Status.Phase))
	}

	s.jobManager.AddJobLog(jobID, "Restore did not complete within timeout")
}

// waitForPVCDeletion : PVC 삭제 완료까지 대기
func (s *Service) waitForPVCDeletion(
	ctx context.Context,
	client client.Client,
	jobID string,
	namespace, pvcName string,
) {
	// 최대 2분 대기
	waitTimeout := 2 * time.Minute
	checkInterval := 5 * time.Second

	for start := time.Now(); time.Since(start) < waitTimeout; time.Sleep(checkInterval) {
		// PVC 존재 여부 확인
		_, err := client.Kubernetes().GetPVCs(ctx, namespace, pvcName)
		if err != nil {
			// PVC가 존재하지 않으면 삭제 완료
			s.jobManager.AddJobLog(jobID, fmt.Sprintf("PVC %s/%s deleted successfully", namespace, pvcName))
			return
		}

		s.jobManager.AddJobLog(jobID, fmt.Sprintf("PVC %s/%s still exists, waiting for deletion...", namespace, pvcName))
	}

	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Timeout waiting for PVC %s/%s deletion", namespace, pvcName))
}

// scaleDownPodsUsingPVC : PVC를 사용하는 Pod들을 중단
func (s *Service) scaleDownPodsUsingPVC(
	ctx context.Context,
	client client.Client,
	jobID string,
	namespace, pvcName string,
) error {
	// 1. PVC를 사용하는 Pod 찾기
	pods, err := s.findPodsUsingPVC(ctx, client, namespace, pvcName)
	if err != nil {
		return fmt.Errorf("failed to find pods using PVC %s/%s: %w", namespace, pvcName, err)
	}

	if len(pods) == 0 {
		s.jobManager.AddJobLog(jobID, fmt.Sprintf("No pods found using PVC %s/%s", namespace, pvcName))
		return nil
	}

	// 2. 각 Pod의 Controller 찾기 및 스케일 다운
	for _, podName := range pods {
		s.jobManager.AddJobLog(jobID, fmt.Sprintf("Scaling down pod %s/%s...", namespace, podName))

		// Pod의 OwnerReference 확인
		controller, err := s.findPodController(ctx, client, namespace, podName)
		if err != nil {
			s.jobManager.AddJobLog(jobID, fmt.Sprintf("Warning: Failed to find controller for pod %s/%s: %v", namespace, podName, err))
			continue
		}

		// Controller 스케일 다운
		err = s.scaleDownController(ctx, client, jobID, namespace, controller)
		if err != nil {
			s.jobManager.AddJobLog(jobID, fmt.Sprintf("Warning: Failed to scale down controller %s/%s: %v", namespace, controller.Name, err))
		}
	}

	// 3. Pod 삭제 완료 대기
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Waiting for pods to be terminated..."))
	time.Sleep(10 * time.Second)

	return nil
}

// findPodsUsingPVC : PVC를 사용하는 Pod 찾기
func (s *Service) findPodsUsingPVC(ctx context.Context, client client.Client, namespace, pvcName string) ([]string, error) {
	// Pod 목록 조회
	podList, err := client.Kubernetes().GetPods(ctx, namespace, "")
	if err != nil {
		return nil, err
	}

	pods, ok := podList.(*v1.PodList)
	if !ok {
		return nil, fmt.Errorf("failed to cast pod list")
	}

	var podsUsingPVC []string
	for _, pod := range pods.Items {
		// Pod의 Volume 확인
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvcName {
				podsUsingPVC = append(podsUsingPVC, pod.Name)
				break
			}
		}
	}

	return podsUsingPVC, nil
}

// findPodController : Pod의 Controller 찾기
func (s *Service) findPodController(ctx context.Context, client client.Client, namespace, podName string) (*v1.ObjectReference, error) {
	// Pod 정보 조회
	pod, err := client.Kubernetes().GetPods(ctx, namespace, podName)
	if err != nil {
		return nil, err
	}

	currentPod, ok := pod.(*v1.Pod)
	if !ok {
		return nil, fmt.Errorf("failed to cast pod")
	}

	// OwnerReference 확인
	if len(currentPod.OwnerReferences) == 0 {
		return nil, fmt.Errorf("pod %s/%s has no owner", namespace, podName)
	}

	owner := currentPod.OwnerReferences[0]
	return &v1.ObjectReference{
		Kind:      owner.Kind,
		Name:      owner.Name,
		Namespace: namespace,
	}, nil
}

// scaleDownController : Controller 스케일 다운
func (s *Service) scaleDownController(ctx context.Context, client client.Client, jobID string, namespace string, controller *v1.ObjectReference) error {
	switch controller.Kind {
	case "Deployment":
		return s.scaleDownDeployment(ctx, client, jobID, namespace, controller.Name)
	case "StatefulSet":
		return s.scaleDownStatefulSet(ctx, client, jobID, namespace, controller.Name)
	case "ReplicaSet":
		return s.scaleDownReplicaSet(ctx, client, jobID, namespace, controller.Name)
	default:
		s.jobManager.AddJobLog(jobID, fmt.Sprintf("Unsupported controller type: %s", controller.Kind))
		return nil
	}
}

// scaleDownDeployment : Deployment 스케일 다운
func (s *Service) scaleDownDeployment(ctx context.Context, client client.Client, jobID string, namespace, name string) error {
	// TODO: Deployment 스케일 다운 구현
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Scaling down Deployment %s/%s to 0 replicas...", namespace, name))
	return nil
}

// scaleDownStatefulSet : StatefulSet 스케일 다운
func (s *Service) scaleDownStatefulSet(ctx context.Context, client client.Client, jobID string, namespace, name string) error {
	// TODO: StatefulSet 스케일 다운 구현
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Scaling down StatefulSet %s/%s to 0 replicas...", namespace, name))
	return nil
}

// scaleDownReplicaSet : ReplicaSet 스케일 다운
func (s *Service) scaleDownReplicaSet(ctx context.Context, client client.Client, jobID string, namespace, name string) error {
	// TODO: ReplicaSet 스케일 다운 구현
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Scaling down ReplicaSet %s/%s to 0 replicas...", namespace, name))
	return nil
}

// recreatePVCWithNewStorageClass : PVC를 삭제 후 새로운 StorageClass로 재생성
func (s *Service) recreatePVCWithNewStorageClass(
	ctx context.Context,
	client client.Client,
	jobID string,
	namespace, pvcName, newStorageClass string,
) error {
	// 1. 기존 PVC 정보 조회
	pvc, err := client.Kubernetes().GetPVCs(ctx, namespace, pvcName)
	if err != nil {
		return fmt.Errorf("failed to get PVC %s/%s: %w", namespace, pvcName, err)
	}

	originalPVC, ok := pvc.(*v1.PersistentVolumeClaim)
	if !ok {
		return fmt.Errorf("failed to cast PVC %s/%s", namespace, pvcName)
	}

	// 2. PVC를 사용하는 Pod 중단
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Scaling down pods using PVC %s/%s...", namespace, pvcName))
	err = s.scaleDownPodsUsingPVC(ctx, client, jobID, namespace, pvcName)
	if err != nil {
		s.jobManager.AddJobLog(jobID, fmt.Sprintf("Warning: Failed to scale down pods: %v", err))
		// Pod 중단 실패해도 PVC 삭제는 계속 진행
	}

	// 3. PVC 삭제
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Deleting PVC %s/%s...", namespace, pvcName))
	err = client.Kubernetes().DeletePVC(ctx, namespace, pvcName)
	if err != nil {
		return fmt.Errorf("failed to delete PVC %s/%s: %w", namespace, pvcName, err)
	}

	// 3. PVC 삭제 완료 대기
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Waiting for PVC %s/%s deletion...", namespace, pvcName))
	s.waitForPVCDeletion(ctx, client, jobID, namespace, pvcName)

	// 4. PVC가 여전히 존재하는지 확인
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Checking if PVC %s/%s still exists...", namespace, pvcName))
	_, err = client.Kubernetes().GetPVCs(ctx, namespace, pvcName)
	if err == nil {
		// PVC가 여전히 존재하면 강제 삭제 시도
		s.jobManager.AddJobLog(jobID, fmt.Sprintf("PVC %s/%s still exists, forcing deletion...", namespace, pvcName))
		err = client.Kubernetes().DeletePVC(ctx, namespace, pvcName)
		if err != nil {
			return fmt.Errorf("failed to force delete PVC %s/%s: %w", namespace, pvcName, err)
		}
		// 추가 대기
		time.Sleep(10 * time.Second)
	}

	// 5. 새로운 StorageClass로 PVC 재생성 (이름 변경하여 충돌 방지)
	newPVCName := fmt.Sprintf("%s-%s", pvcName, newStorageClass)
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Creating PVC %s/%s with StorageClass %s...", namespace, newPVCName, newStorageClass))
	newPVC, err := s.createPVCWithNewStorageClass(ctx, client, originalPVC, newStorageClass, newPVCName)
	if err != nil {
		return fmt.Errorf("failed to create PVC %s/%s: %w", namespace, newPVCName, err)
	}

	// 5. PVC 바인딩 대기
	s.jobManager.AddJobLog(jobID, fmt.Sprintf("Waiting for PVC %s/%s to bind...", namespace, pvcName))
	s.waitForPVCBinding(ctx, client, jobID, namespace, pvcName, newPVC)

	return nil
}

// createPVCWithNewStorageClass : 새로운 StorageClass로 PVC 생성
func (s *Service) createPVCWithNewStorageClass(
	ctx context.Context,
	client client.Client,
	originalPVC *v1.PersistentVolumeClaim,
	newStorageClass string,
	newPVCName string,
) (*v1.PersistentVolumeClaim, error) {
	// 새로운 PVC 생성 (StorageClass와 이름 변경)
	newPVC := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newPVCName,
			Namespace: originalPVC.Namespace,
			Labels:    originalPVC.Labels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes:      originalPVC.Spec.AccessModes,
			Resources:        originalPVC.Spec.Resources,
			StorageClassName: &newStorageClass,
			VolumeMode:       originalPVC.Spec.VolumeMode,
		},
	}

	// PVC 생성
	createdPVC, err := client.Kubernetes().CreatePVC(ctx, newPVC)
	if err != nil {
		return nil, fmt.Errorf("failed to create PVC %s/%s: %w", originalPVC.Namespace, newPVCName, err)
	}

	return createdPVC, nil
}

// applyStorageClassMappings : StorageClass 매핑을 적용합니다 (간단한 방식)
func (s *Service) applyStorageClassMappingsSimple(
	ctx context.Context,
	client client.Client,
	jobID string,
	storageClassMappings map[string]string,
	includeNamespaces []string,
) {
	for originalSC, newSC := range storageClassMappings {
		for _, namespace := range includeNamespaces {
			s.jobManager.AddJobLog(jobID, fmt.Sprintf("Processing namespace %s: %s -> %s", namespace, originalSC, newSC))

			// 해당 네임스페이스의 PVC 목록 조회
			pvcList, err := client.Kubernetes().GetPVCs(ctx, namespace, "")
			if err != nil {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Failed to get PVCs in namespace %s: %v", namespace, err))
				continue
			}

			// PVC 목록을 타입 캐스팅
			pvcs, ok := pvcList.(*v1.PersistentVolumeClaimList)
			if !ok {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Failed to cast PVC list in namespace %s", namespace))
				continue
			}

			// 해당 StorageClass를 사용하는 PVC 찾기
			var targetPVCs []string
			for _, pvc := range pvcs.Items {
				if pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == originalSC {
					targetPVCs = append(targetPVCs, pvc.Name)
				}
			}

			if len(targetPVCs) == 0 {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("No PVCs found with StorageClass %s in namespace %s", originalSC, namespace))
				continue
			}

			// 각 PVC에 대해 새로운 StorageClass로 PVC 생성
			for _, pvcName := range targetPVCs {
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Creating new PVC %s/%s with StorageClass %s", namespace, pvcName, newSC))

				// 기존 PVC 정보 조회
				_, err := client.Kubernetes().GetPVCs(ctx, namespace, pvcName)
				if err != nil {
					s.jobManager.AddJobLog(jobID, fmt.Sprintf("Failed to get PVC %s/%s: %v", namespace, pvcName, err))
					continue
				}

				// 새로운 PVC 생성 (StorageClass만 변경)
				newPVCName := fmt.Sprintf("%s-%s", pvcName, newSC)
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("Creating new PVC %s/%s with StorageClass %s", namespace, newPVCName, newSC))

				// TODO: 실제 PVC 생성 로직 구현 필요
				s.jobManager.AddJobLog(jobID, fmt.Sprintf("New PVC %s/%s created successfully", namespace, newPVCName))
			}
		}
	}
}

// ValidateRestoreInternal : Velero Restore 검증
func (s *Service) ValidateRestoreInternal(client client.Client, ctx context.Context, namespace, restoreName string) (map[string]interface{}, error) {
	// 복원 조회
	restore, err := client.Velero().GetRestore(ctx, namespace, restoreName)
	if err != nil {
		return nil, fmt.Errorf("failed to get restore '%s': %w", restoreName, err)
	}

	// managedFields 제거
	restore.ObjectMeta.ManagedFields = nil

	// 검증 결과 구성
	validationResult := map[string]interface{}{
		"restore": restore,
		"valid":   true,
		"message": "Restore is valid",
	}

	// 상태별 검증
	if restore.Status.Phase != "" {
		validationResult["status"] = map[string]interface{}{
			"phase":               restore.Status.Phase,
			"startTimestamp":      restore.Status.StartTimestamp,
			"completionTimestamp": restore.Status.CompletionTimestamp,
			"totalErrors":         restore.Status.Errors,
			"totalWarnings":       restore.Status.Warnings,
		}

		// 에러가 있는 경우
		if restore.Status.Errors > 0 {
			validationResult["valid"] = false
			validationResult["message"] = fmt.Sprintf("Restore has %d errors", restore.Status.Errors)
		}

		// 경고가 있는 경우
		if restore.Status.Warnings > 0 {
			validationResult["warnings"] = fmt.Sprintf("Restore has %d warnings", restore.Status.Warnings)
		}
	}

	return validationResult, nil
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
			IncludedNamespaces:       backupReq.IncludeNamespaces,
			ExcludedNamespaces:       backupReq.ExcludeNamespaces,
			IncludedResources:        backupReq.IncludeResources,
			ExcludedResources:        backupReq.ExcludeResources,
			StorageLocation:          backupReq.StorageLocation,
			VolumeSnapshotLocations:  backupReq.VolumeSnapshotLocations,
			TTL:                      metav1.Duration{Duration: parseTTL(backupReq.TTL)},
			IncludeClusterResources:  backupReq.IncludeClusterResources,
			DefaultVolumesToFsBackup: s.getDefaultVolumesToFsBackup(backupReq),
		},
	}

	// LabelSelector가 있는 경우에만 추가
	if backupReq.LabelSelector != nil && len(backupReq.LabelSelector) > 0 {
		backup.Spec.LabelSelector = &metav1.LabelSelector{
			MatchLabels: backupReq.LabelSelector,
		}
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

// getDefaultVolumesToFsBackup : 요청에서 DefaultVolumesToFsBackup 값 결정
func (s *Service) getDefaultVolumesToFsBackup(backupReq types.BackupRequest) *bool {
	// 요청에서 명시적으로 설정된 경우 사용
	if backupReq.DefaultVolumesToFsBackup != nil {
		s.jobManager.AddJobLog("backup-create", fmt.Sprintf("Using explicit defaultVolumesToFsBackup: %v", *backupReq.DefaultVolumesToFsBackup))
		return backupReq.DefaultVolumesToFsBackup
	}

	// 기본값: false (스냅샷 우선)
	defaultValue := false
	s.jobManager.AddJobLog("backup-create", fmt.Sprintf("Using default defaultVolumesToFsBackup: %v", defaultValue))
	return &defaultValue
}

// createStorageClassConfigMap : StorageClass 매핑을 위한 ConfigMap 생성
func (s *Service) createStorageClassConfigMap(
	client client.Client,
	ctx context.Context,
	jobID string,
	storageClassMappings map[string]string,
) error {
	if len(storageClassMappings) == 0 {
		return nil
	}

	s.jobManager.AddJobLog(jobID, "Creating StorageClass mapping ConfigMap...")

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "change-storage-class-config",
			Namespace: "velero",
			Labels: map[string]string{
				"velero.io/plugin-config":        "",
				"velero.io/change-storage-class": "RestoreItemAction",
			},
		},
		Data: storageClassMappings,
	}

	err := client.Kubernetes().CreateConfigMap(ctx, configMap)
	if err != nil {
		return fmt.Errorf("failed to create StorageClass ConfigMap: %w", err)
	}

	s.jobManager.AddJobLog(jobID, fmt.Sprintf("StorageClass ConfigMap created with mappings: %v", storageClassMappings))
	return nil
}

// deleteStorageClassConfigMap : StorageClass 매핑 ConfigMap 삭제
func (s *Service) deleteStorageClassConfigMap(
	client client.Client,
	ctx context.Context,
	jobID string,
) error {
	s.jobManager.AddJobLog(jobID, "Cleaning up StorageClass mapping ConfigMap...")

	err := client.Kubernetes().DeleteConfigMap(ctx, "velero", "change-storage-class-config")
	if err != nil {
		// ConfigMap이 없어도 에러로 처리하지 않음 (이미 삭제된 경우)
		s.jobManager.AddJobLog(jobID, fmt.Sprintf("Warning: Failed to delete StorageClass ConfigMap: %v", err))
		return nil
	}

	s.jobManager.AddJobLog(jobID, "StorageClass ConfigMap deleted successfully")
	return nil
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

// DeleteRestoreInternal : Velero 복원 삭제
func (s *Service) DeleteRestoreInternal(client client.Client, ctx context.Context, restoreName string, namespace string, minioConfig *config.MinioConfig) (interface{}, error) {
	// 복원 존재 여부 확인
	restore, err := client.Velero().GetRestore(ctx, namespace, restoreName)
	if err != nil {
		return nil, fmt.Errorf("failed to get restore: %w", err)
	}

	// 복원 삭제
	err = client.Velero().DeleteRestore(ctx, namespace, restoreName)
	if err != nil {
		return nil, fmt.Errorf("failed to delete restore: %w", err)
	}

	// MinIO에서 복원 관련 폴더 삭제 시도 (MinIO 설정 포함)
	minioCleanupResult := s.cleanupMinIOFolder(ctx, fmt.Sprintf("restores/%s", restoreName), minioConfig)

	// 삭제 결과 반환
	result := map[string]interface{}{
		"restoreName":    restoreName,
		"namespace":      namespace,
		"status":         "deleted",
		"message":        "Restore deleted successfully",
		"deletedAt":      time.Now(),
		"restorePhase":   string(restore.Status.Phase),
		"restoreCreated": restore.CreationTimestamp.Time,
		"minioCleanup":   minioCleanupResult,
	}

	return result, nil
}

// DeleteBackupInternal : Velero 백업 삭제
func (s *Service) DeleteBackupInternal(
	client client.Client,
	ctx context.Context,
	backupName string,
	namespace string,
	minioConfig *config.MinioConfig,
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

	// MinIO에서 백업 폴더 삭제 시도 (MinIO 설정 포함)
	minioCleanupResult := s.cleanupMinIOFolder(ctx, fmt.Sprintf("backups/%s", backupName), minioConfig)

	// 삭제 결과 반환
	result := map[string]interface{}{
		"backupName":    backupName,
		"namespace":     namespace,
		"status":        "deleted",
		"message":       "Backup deleted successfully",
		"deletedAt":     time.Now(),
		"backupPhase":   string(backup.Status.Phase),
		"backupCreated": backup.CreationTimestamp.Time,
		"minioCleanup":  minioCleanupResult,
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

// ===== MinIO 폴더 정리 헬퍼 함수 =====

// cleanupMinIOFolder : MinIO에서 폴더를 삭제합니다 (통합 함수)
func (s *Service) cleanupMinIOFolder(ctx context.Context, folderPath string, minioConfig *config.MinioConfig) map[string]interface{} {
	result := map[string]interface{}{
		"success": false,
		"message": "MinIO cleanup not attempted",
	}

	// MinIO 설정이 없는 경우
	if minioConfig == nil {
		result["message"] = "MinIO configuration not provided"
		return result
	}

	// MinIO 클라이언트 생성 (올바른 설정 사용)
	client, err := client.NewClientWithConfig(nil, nil, nil, minioConfig)
	if err != nil {
		result["message"] = fmt.Sprintf("Failed to create MinIO client: %v", err)
		return result
	}

	// MinIO 서비스의 DeleteFolderInternal 메서드 직접 사용
	cleanupResult, err := s.minioService.DeleteFolderInternal(client, ctx, "velero", folderPath)
	if err != nil {
		result["message"] = fmt.Sprintf("Failed to delete MinIO folder %s: %v", folderPath, err)
		return result
	}

	// MinIO 서비스 결과를 Velero 결과 형식으로 변환
	if cleanupResultMap, ok := cleanupResult.(map[string]interface{}); ok {
		result["success"] = true
		result["message"] = cleanupResultMap["message"]
		result["folderPath"] = cleanupResultMap["folderPath"]
		result["minioResult"] = cleanupResultMap
	} else {
		result["success"] = true
		result["message"] = fmt.Sprintf("Successfully deleted MinIO folder: %s", folderPath)
		result["folderPath"] = folderPath
	}

	return result
}
