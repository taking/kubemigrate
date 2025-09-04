package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/pkg/cache"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/health"
	_ "github.com/taking/kubemigrate/pkg/models"
	"github.com/taking/kubemigrate/pkg/response"
	"github.com/taking/kubemigrate/pkg/utils"
	"github.com/taking/kubemigrate/pkg/validator"
)

// MinioHandler : Minio 관련 HTTP 요청을 처리하는 핸들러
type MinioHandler struct {
	minioValidator *validator.MinioValidator
	cache          *cache.Cache
	workerPool     *utils.WorkerPool
	healthManager  *health.HealthManager
}

// NewMinioHandler : 새로운 MinioHandler 인스턴스 생성
func NewMinioHandler(appCache *cache.Cache, workerPool *utils.WorkerPool, healthManager *health.HealthManager) *MinioHandler {
	return &MinioHandler{
		minioValidator: validator.NewMinioValidator(),
		cache:          appCache,
		workerPool:     workerPool,
		healthManager:  healthManager,
	}
}

// HealthCheck : Minio 연결 상태 확인
// @Summary Minio 연결 상태 확인
// @Description Validate Minio connection using MinioConfig
// @Tags minio
// @Accept json
// @Produce json
// @Param request body models.MinioConfig true "Minio connection configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Connection successful"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Failure 503 {object} models.SwaggerErrorResponse "Service unavailable"
// @Router /minio/health [get]
func (h *MinioHandler) HealthCheck(c echo.Context) error {
	req, err := utils.BindAndValidateMinioConfig(c, h.minioValidator)
	if err != nil {
		return err
	}

	client, err := client.NewMinioClient(req)
	if err != nil {
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}

	// 클러스터 연결 상태 확인
	if err := client.HealthCheck(c.Request().Context()); err != nil {
		return response.RespondError(c, http.StatusServiceUnavailable,
			"Minio cluster unhealthy: "+err.Error())
	}

	// 헬스체크 매니저에 MinIO 체커 등록
	if h.healthManager != nil {
		minioChecker := health.NewMinioHealthChecker(req)
		h.healthManager.Register(minioChecker)
	}

	return response.RespondStatus(c, "healthy", "Minio connection successful")
}

// CreateBucketIfNotExists : Minio 버킷 생성 여부 확인
// @Summary Minio 버킷 생성 여부 확인
// @Description Create Minio bucket if it doesn't exist using MinioConfig
// @Tags minio
// @Accept json
// @Produce json
// @Param request body models.MinioConfig true "Minio configuration with bucket name"
// @Success 200 {object} models.SwaggerSuccessResponse "Bucket created or already exists"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /minio/bucket_check [post]
func (h *MinioHandler) CreateBucketIfNotExists(c echo.Context) error {
	req, err := utils.BindAndValidateMinioConfig(c, h.minioValidator)
	if err != nil {
		return err
	}

	client, err := client.NewMinioClient(req)
	if err != nil {
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}

	exists, err := client.BucketExists(c.Request().Context(), req.BucketName)
	if err != nil {
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}

	if !exists {
		if err := client.CreateBucket(c.Request().Context(), req.BucketName); err != nil {
			return response.RespondError(c, http.StatusInternalServerError, err.Error())
		}
		return response.RespondStatus(c, "created", "Bucket created successfully")
	}

	return response.RespondStatus(c, "exists", "Bucket already exists")
}
