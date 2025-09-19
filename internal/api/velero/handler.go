package velero

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/utils"
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
	namespace := utils.ResolveNamespace(c, "velero")
	force := utils.ResolveBool(c, "force", false)

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
	unifiedClient := client.NewClientWithConfig(
		&config.KubeConfig,
		&config.KubeConfig,
		&config,
		&config.MinioConfig,
	)

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
		namespace := utils.ResolveNamespace(c, "velero")
		return h.service.GetBackupsInternal(client, ctx, namespace)
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
		namespace := utils.ResolveNamespace(c, "velero")
		return h.service.GetRestoresInternal(client, ctx, namespace)
	})
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
		namespace := utils.ResolveNamespace(c, "velero")
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
		namespace := utils.ResolveNamespace(c, "velero")
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
		namespace := utils.ResolveNamespace(c, "velero")
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
		namespace := utils.ResolveNamespace(c, "velero")
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
// @Router /api/v1/velero/status/{jobId} [get]
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
// @Router /api/v1/velero/logs/{jobId} [get]
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
// @Router /api/v1/velero/jobs [get]
func (h *Handler) GetAllJobs(c echo.Context) error {
	result, err := h.service.GetAllJobsInternal()
	if err != nil {
		return response.RespondWithErrorModel(c, 500, "INTERNAL_ERROR", err.Error(), "")
	}

	return response.RespondWithData(c, 200, result)
}
