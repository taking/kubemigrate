package velero

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/config"
	"github.com/taking/kubemigrate/pkg/types"
)

// Handler : Velero 관련 HTTP 핸들러
type Handler struct {
	*handler.BaseHandler
	service *Service
}

// NewHandler : 새로운 Velero 핸들러 생성
func NewHandler(base *handler.BaseHandler) *Handler {
	return &Handler{
		BaseHandler: base,
		service:     NewService(base),
	}
}

// HealthCheck : Velero 연결 테스트
// @Summary Velero Connection Test
// @Description Test Velero connection with provided configuration
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/health [post]
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.BaseHandler.HealthCheck(c, handler.HealthCheckConfig{
		ServiceName: "velero",
		DefaultNS:   "", // Velero는 고정 네임스페이스 사용
		HealthFunc: func(client client.Client, ctx context.Context) error {
			_, err := client.Velero().GetBackups(ctx, "velero")
			return err
		},
	})
}

// InstallVeleroWithMinIO : Velero 설치 및 MinIO 연동 설정
// @Summary Install Velero with MinIO
// @Description Install Velero and configure MinIO integration
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Param namespace query string false "Namespace name (default: 'velero')"
// @Param force query boolean false "Force recreate BSL and MinIO Secret (default: false)"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/install [post]
func (h *Handler) InstallVeleroWithMinIO(c echo.Context) error {
	// 공통 검증 및 바인딩 사용
	config, err := h.ValidateVeleroConfig(c, "velero")
	if err != nil {
		return err
	}

	// Query 파라미터 처리
	namespace := h.ResolveNamespace(c, "velero")
	force := h.ResolveBool(c, "force", false)

	// 컨텍스트 생성 (타임아웃 설정)
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Minute)
	defer cancel()

	// MinIO 클라이언트 직접 생성 및 테스트
	minioClient, err := minio.NewClientWithConfig(config.MinioConfig)
	if err != nil {
		return h.HandleConnectionError(c, "velero", "minio client creation", err)
	}

	// MinIO 연결 테스트
	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		return h.HandleConnectionError(c, "velero", "minio connection", err)
	}

	// 통합 클라이언트 생성 (MinIO 클라이언트는 이미 검증됨)
	unifiedClient, err := client.NewClientWithConfig(
		&config.KubeConfig,
		&config.KubeConfig,
		&config,
		&config.MinioConfig,
	)
	if err != nil {
		return h.HandleConnectionError(c, "velero", "unified client creation", err)
	}

	// Velero 설치 및 MinIO 연동 실행
	result, err := h.service.InstallVeleroWithMinIOInternal(unifiedClient, ctx, config, namespace, force)
	if err != nil {
		return h.HandleInternalError(c, "velero", "installation", err)
	}

	return response.RespondWithData(c, 200, result)
}

// GetBackups : Velero 백업 목록 조회
// @Summary Get Velero Backups
// @Description Get list of Velero backups
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param namespace query string false "Namespace name (default: 'velero')"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/backups [get]
func (h *Handler) GetBackups(c echo.Context) error {
	return h.HandleResourceClient(c, "velero-backups", func(client client.Client, ctx context.Context) (interface{}, error) {
		namespace := h.ResolveNamespace(c, "velero")
		return h.service.GetBackupsInternal(client, ctx, namespace)
	})
}

// GetBackup : Velero 백업 상세 조회
// @Summary Get Velero Backup Details
// @Description Get detailed information about a specific Velero backup
// @Tags velero
// @Accept json
// @Produce json
// @Param backupName path string true "Backup name"
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param namespace query string false "Namespace name (default: 'velero')"
// @Success 200 {object} response.SuccessResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/backups/{backupName} [get]
func (h *Handler) GetBackup(c echo.Context) error {
	backupName := c.Param("backupName")
	if backupName == "" {
		return response.RespondWithErrorModel(c, 400, "INVALID_REQUEST", "Backup name is required", "")
	}

	return h.HandleResourceClient(c, "velero-backup", func(client client.Client, ctx context.Context) (interface{}, error) {
		namespace := h.ResolveNamespace(c, "velero")
		return h.service.GetBackupInternal(client, ctx, namespace, backupName)
	})
}

// GetRestores : Velero 복원 목록 조회
// @Summary Get Velero Restores
// @Description Get list of Velero restores
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param namespace query string false "Namespace name (default: 'velero')"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/restores [get]
func (h *Handler) GetRestores(c echo.Context) error {
	return h.HandleResourceClient(c, "velero-restores", func(client client.Client, ctx context.Context) (interface{}, error) {
		namespace := h.ResolveNamespace(c, "velero")
		return h.service.GetRestoresInternal(client, ctx, namespace)
	})
}

// GetRestore : Velero 복원 상세 조회
// @Summary Get Velero Restore Details
// @Description Get detailed information about a specific Velero restore
// @Tags velero
// @Accept json
// @Produce json
// @Param restoreName path string true "Restore name"
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param namespace query string false "Namespace name (default: 'velero')"
// @Success 200 {object} response.SuccessResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/restores/{restoreName} [get]
func (h *Handler) GetRestore(c echo.Context) error {
	restoreName := c.Param("restoreName")
	if restoreName == "" {
		return response.RespondWithErrorModel(c, 400, "INVALID_REQUEST", "Restore name is required", "")
	}

	return h.HandleResourceClient(c, "velero-restore", func(client client.Client, ctx context.Context) (interface{}, error) {
		namespace := h.ResolveNamespace(c, "velero")
		return h.service.GetRestoreInternal(client, ctx, namespace, restoreName)
	})
}

// CreateRestore : Velero 복원 생성
// @Summary Create Velero Restore
// @Description Create a new Velero restore from a backup
// @Tags velero
// @Accept json
// @Produce json
// @Param request body types.VeleroRestoreRequest true "Restore configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/restores [post]
func (h *Handler) CreateRestore(c echo.Context) error {
	var req types.VeleroRestoreRequest
	if err := c.Bind(&req); err != nil {
		return response.RespondWithErrorModel(c, 400, "INVALID_REQUEST_BODY", "Invalid request body format", err.Error())
	}

	// 필수 필드 검증
	if req.Restore.Name == "" {
		return response.RespondWithErrorModel(c, 400, "INVALID_REQUEST", "Restore name is required", "")
	}
	if req.Restore.BackupName == "" {
		return response.RespondWithErrorModel(c, 400, "INVALID_REQUEST", "Backup name is required", "")
	}

	// 클라이언트 생성
	client, err := client.NewClientWithConfig(req.KubeConfig, "", "", "")
	if err != nil {
		return response.RespondWithErrorModel(c, 500, "CLIENT_CREATION_FAILED", "Failed to create client", err.Error())
	}

	// 복원 생성
	result, err := h.service.CreateRestoreInternal(client, context.Background(), req)
	if err != nil {
		return response.RespondWithErrorModel(c, 500, "RESTORE_CREATION_FAILED", "Failed to create restore", err.Error())
	}

	return response.RespondWithData(c, 200, result)
}

// ValidateRestore : Velero 복원 검증
// @Summary Validate Velero Restore
// @Description Validate a specific Velero restore
// @Tags velero
// @Accept json
// @Produce json
// @Param restoreName path string true "Restore name"
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param namespace query string false "Namespace name (default: 'velero')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/restores/{restoreName}/validate [post]
func (h *Handler) ValidateRestore(c echo.Context) error {
	restoreName := c.Param("restoreName")
	if restoreName == "" {
		return response.RespondWithErrorModel(c, 400, "INVALID_REQUEST", "Restore name is required", "")
	}

	return h.HandleResourceClient(c, "velero-restore-validation", func(client client.Client, ctx context.Context) (interface{}, error) {
		namespace := h.ResolveNamespace(c, "velero")
		return h.service.ValidateRestoreInternal(client, ctx, namespace, restoreName)
	})
}

// DeleteRestore : Velero 복원 삭제
// @Summary Delete Velero Restore
// @Description Delete a specific Velero restore
// @Tags velero
// @Accept json
// @Produce json
// @Param request body types.DeleteRestoreRequest true "Delete restore request with kubeconfig and minio config"
// @Param restoreName path string true "Restore name"
// @Param namespace query string false "Namespace name (default: 'velero')"
// @Success 200 {object} map[string]interface{} "Restore deletion result"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/restores/{restoreName} [delete]
func (h *Handler) DeleteRestore(c echo.Context) error {
	// 복원 이름 파라미터 추출
	restoreName := c.Param("restoreName")
	if restoreName == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "restoreName is required", "")
	}

	// 삭제 요청 바인딩
	var deleteReq types.DeleteRestoreRequest
	if err := c.Bind(&deleteReq); err != nil {
		return response.RespondWithErrorModel(c, 400, "INVALID_REQUEST", "Invalid request body format", "")
	}

	// 필수 필드 검증
	if deleteReq.KubeConfig.KubeConfig == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "kubeconfig is required", "")
	}
	if deleteReq.MinioConfig.Endpoint == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "minio endpoint is required", "")
	}
	if deleteReq.MinioConfig.AccessKey == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "minio accessKey is required", "")
	}
	if deleteReq.MinioConfig.SecretKey == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "minio secretKey is required", "")
	}

	// 컨텍스트 생성
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Minute)
	defer cancel()

	// 클라이언트 생성
	unifiedClient, err := client.NewClientWithConfig(
		&deleteReq.KubeConfig,
		&deleteReq.KubeConfig,
		nil,
		&deleteReq.MinioConfig,
	)
	if err != nil {
		return h.HandleConnectionError(c, "velero", "client creation", err)
	}

	// 복원 삭제 실행 (MinIO 설정 포함)
	namespace := h.ResolveNamespace(c, "velero")
	result, err := h.service.DeleteRestoreInternal(unifiedClient, ctx, restoreName, namespace, &deleteReq.MinioConfig)
	if err != nil {
		return h.HandleInternalError(c, "velero", "restore deletion", err)
	}

	return response.RespondWithData(c, 200, result)
}

// GetBackupRepositories : 백업 저장소 조회
// @Summary Get Backup Repositories
// @Description Get list of backup repositories
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/repositories [get]
func (h *Handler) GetBackupRepositories(c echo.Context) error {
	return h.HandleResourceClient(c, "velero-repositories", func(client client.Client, ctx context.Context) (interface{}, error) {
		// Velero 백업 저장소 조회
		namespace := h.ResolveNamespace(c, "velero")
		return h.service.GetBackupRepositoriesInternal(client, ctx, namespace)
	})
}

// GetBackupStorageLocations : 백업 스토리지 위치 조회
// @Summary Get Backup Storage Locations
// @Description Get list of backup storage locations
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/storage-locations [get]
func (h *Handler) GetBackupStorageLocations(c echo.Context) error {
	return h.HandleResourceClient(c, "velero-storage-locations", func(client client.Client, ctx context.Context) (interface{}, error) {
		// Velero 백업 스토리지 위치 조회
		namespace := h.ResolveNamespace(c, "velero")
		return h.service.GetBackupStorageLocationsInternal(client, ctx, namespace)
	})
}

// GetVolumeSnapshotLocations : 볼륨 스냅샷 위치 조회
// @Summary Get Volume Snapshot Locations
// @Description Get list of volume snapshot locations
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/volume-snapshot-locations [get]
func (h *Handler) GetVolumeSnapshotLocations(c echo.Context) error {
	return h.HandleResourceClient(c, "velero-volume-snapshot-locations", func(client client.Client, ctx context.Context) (interface{}, error) {
		// Velero 볼륨 스냅샷 위치 조회
		namespace := h.ResolveNamespace(c, "velero")
		return h.service.GetVolumeSnapshotLocationsInternal(client, ctx, namespace)
	})
}

// GetPodVolumeRestores : Pod 볼륨 복원 조회
// @Summary Get Pod Volume Restores
// @Description Get list of pod volume restores
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/pod-volume-restores [get]
func (h *Handler) GetPodVolumeRestores(c echo.Context) error {
	return h.HandleResourceClient(c, "velero-pod-volume-restores", func(client client.Client, ctx context.Context) (interface{}, error) {
		// Velero Pod 볼륨 복원 조회
		namespace := h.ResolveNamespace(c, "velero")
		return h.service.GetPodVolumeRestoresInternal(client, ctx, namespace)
	})
}

// GetJobStatus : 작업 상태 조회
// @Summary Get Job Status
// @Description Get the status of a specific job
// @Tags velero
// @Accept json
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} map[string]interface{} "Job status"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/velero/status/{jobId} [get]
func (h *Handler) GetJobStatus(c echo.Context) error {
	jobID := c.Param("jobId")
	if jobID == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "jobId is required", "")
	}

	result, err := h.service.GetJobStatusInternal(jobID)
	if err != nil {
		return response.RespondWithErrorModel(c, 404, "JOB_NOT_FOUND", err.Error(), "")
	}

	return response.RespondWithData(c, 200, result)
}

// GetJobLogs : 작업 로그 조회
// @Summary Get Job Logs
// @Description Get the logs of a specific job
// @Tags velero
// @Accept json
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} map[string]interface{} "Job logs"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/velero/logs/{jobId} [get]
func (h *Handler) GetJobLogs(c echo.Context) error {
	jobID := c.Param("jobId")
	if jobID == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "jobId is required", "")
	}

	result, err := h.service.GetJobLogsInternal(jobID)
	if err != nil {
		return response.RespondWithErrorModel(c, 404, "JOB_NOT_FOUND", err.Error(), "")
	}

	return response.RespondWithData(c, 200, result)
}

// GetAllJobs : 모든 작업 조회
// @Summary Get All Jobs
// @Description Get all jobs
// @Tags velero
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "All jobs"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /v1/velero/jobs [get]
func (h *Handler) GetAllJobs(c echo.Context) error {
	result, err := h.service.GetAllJobsInternal()
	if err != nil {
		return response.RespondWithErrorModel(c, 500, "INTERNAL_ERROR", err.Error(), "")
	}

	return response.RespondWithData(c, 200, result)
}

// CreateBackup : Velero 백업 생성
// @Summary Create Velero Backup
// @Description Create a new Velero backup with specified configuration
// @Tags velero
// @Accept json
// @Produce json
// @Param request body types.CreateBackupRequest true "Backup configuration with kubeconfig and minio settings"
// @Param namespace query string false "Namespace name (default: 'velero')"
// @Success 200 {object} types.BackupResult
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/backups [post]
func (h *Handler) CreateBackup(c echo.Context) error {
	// 백업 요청 바인딩
	var createBackupReq types.CreateBackupRequest
	if err := c.Bind(&createBackupReq); err != nil {
		// 에러 메시지를 간단하게 처리
		errorMsg := "Invalid request body format"
		if err.Error() != "" {
			errorMsg = "Request body validation failed"
		}
		return response.RespondWithErrorModel(c, 400, "INVALID_REQUEST_BODY", errorMsg, "")
	}

	// 필수 필드 검증
	if createBackupReq.Backup.Name == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "backup name is required", "")
	}

	// kubeconfig 검증
	if createBackupReq.KubeConfig.KubeConfig == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "kubeconfig is required", "")
	}

	// minio 설정 검증
	if createBackupReq.MinioConfig.Endpoint == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "minio endpoint is required", "")
	}
	if createBackupReq.MinioConfig.AccessKey == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "minio accessKey is required", "")
	}
	if createBackupReq.MinioConfig.SecretKey == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "minio secretKey is required", "")
	}

	// Query 파라미터 처리
	namespace := h.ResolveNamespace(c, "velero")

	// Velero 설정 구성
	veleroConfig := config.VeleroConfig{
		KubeConfig:  createBackupReq.KubeConfig,
		MinioConfig: createBackupReq.MinioConfig,
	}

	// 컨텍스트 생성 (타임아웃 설정)
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Minute)
	defer cancel()

	// 클라이언트 생성
	unifiedClient, err := client.NewClientWithConfig(
		&veleroConfig.KubeConfig,
		&veleroConfig.KubeConfig,
		&veleroConfig,
		&veleroConfig.MinioConfig,
	)
	if err != nil {
		return h.HandleConnectionError(c, "velero", "client creation", err)
	}

	// 백업 생성 실행
	result, err := h.service.CreateBackupInternal(unifiedClient, ctx, createBackupReq.Backup, namespace)
	if err != nil {
		return h.HandleInternalError(c, "velero", "backup creation", err)
	}

	return response.RespondWithData(c, 200, result)
}

// ValidateBackup : Velero 백업 검증
// @Summary Validate Velero Backup
// @Description Validate a specific Velero backup and check its status
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param backupName path string true "Backup name"
// @Param namespace query string false "Namespace name (default: 'velero')"
// @Success 200 {object} types.BackupValidationResult
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/backups/{backupName}/validate [post]
func (h *Handler) ValidateBackup(c echo.Context) error {
	// 백업 이름 파라미터 추출
	backupName := c.Param("backupName")
	if backupName == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "backupName is required", "")
	}

	return h.HandleResourceClient(c, "velero-backup-validation", func(client client.Client, ctx context.Context) (interface{}, error) {
		namespace := h.ResolveNamespace(c, "velero")
		return h.service.ValidateBackupInternal(client, ctx, backupName, namespace)
	})
}

// DeleteBackup : Velero 백업 삭제
// @Summary Delete Velero Backup
// @Description Delete a specific Velero backup
// @Tags velero
// @Accept json
// @Produce json
// @Param request body types.DeleteBackupRequest true "Delete backup request with kubeconfig and minio config"
// @Param backupName path string true "Backup name"
// @Param namespace query string false "Namespace name (default: 'velero')"
// @Success 200 {object} map[string]interface{} "Backup deletion result"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/backups/{backupName} [delete]
func (h *Handler) DeleteBackup(c echo.Context) error {
	// 백업 이름 파라미터 추출
	backupName := c.Param("backupName")
	if backupName == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "backupName is required", "")
	}

	// 삭제 요청 바인딩
	var deleteReq types.DeleteBackupRequest
	if err := c.Bind(&deleteReq); err != nil {
		return response.RespondWithErrorModel(c, 400, "INVALID_REQUEST", "Invalid request body format", "")
	}

	// 필수 필드 검증
	if deleteReq.KubeConfig.KubeConfig == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "kubeconfig is required", "")
	}
	if deleteReq.MinioConfig.Endpoint == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "minio endpoint is required", "")
	}
	if deleteReq.MinioConfig.AccessKey == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "minio accessKey is required", "")
	}
	if deleteReq.MinioConfig.SecretKey == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "minio secretKey is required", "")
	}

	// 컨텍스트 생성
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Minute)
	defer cancel()

	// 클라이언트 생성
	unifiedClient, err := client.NewClientWithConfig(
		&deleteReq.KubeConfig,
		&deleteReq.KubeConfig,
		nil,
		&deleteReq.MinioConfig,
	)
	if err != nil {
		return h.HandleConnectionError(c, "velero", "client creation", err)
	}

	// 백업 삭제 실행 (MinIO 설정 포함)
	namespace := h.ResolveNamespace(c, "velero")
	result, err := h.service.DeleteBackupInternal(unifiedClient, ctx, backupName, namespace, &deleteReq.MinioConfig)
	if err != nil {
		return h.HandleInternalError(c, "velero", "backup deletion", err)
	}

	return response.RespondWithData(c, 200, result)
}
