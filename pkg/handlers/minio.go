package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"taking.kr/velero/pkg/cache"
	"taking.kr/velero/pkg/client"
	"taking.kr/velero/pkg/health"
	"taking.kr/velero/pkg/interfaces"
	"taking.kr/velero/pkg/models"
	"taking.kr/velero/pkg/response"
	"taking.kr/velero/pkg/utils"
	"taking.kr/velero/pkg/validator"
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
	req, err := h.bindAndValidateMinioConfig(c)
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
	req, err := h.bindAndValidateMinioConfig(c)
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

// handleMinioResourceWithCache : 캐시를 사용하는 Minio 리소스 처리 헬퍼
func (h *MinioHandler) handleMinioResourceWithCache(c echo.Context, cacheKey string,
	getResource func(interfaces.MinioClient) (interface{}, error)) error {

	// MinioConfig 검증
	req, err := h.bindAndValidateMinioConfig(c)
	if err != nil {
		return err
	}

	// 캐시 키 생성
	fullCacheKey := fmt.Sprintf("minio:%s:%s", cacheKey, req.Endpoint)

	// 캐시에서 가져오기
	if cached, exists := h.cache.Get(fullCacheKey); exists {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   cached,
			"cached": true,
		})
	}

	// 캐시에 없으면 클라이언트에서 가져오기
	minioClient, err := client.NewMinioClient(req)
	if err != nil {
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}

	// 워커 풀을 사용하여 백그라운드에서 데이터 가져오기
	resultChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)

	h.workerPool.Submit(func() {
		data, err := getResource(minioClient)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- data
	})

	select {
	case data := <-resultChan:
		// 캐시에 저장
		h.cache.Set(fullCacheKey, data)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   data,
			"cached": false,
		})
	case err := <-errorChan:
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}
}

// bindAndValidateMinioConfig : MinioConfig 검증
func (h *MinioHandler) bindAndValidateMinioConfig(c echo.Context) (models.MinioConfig, error) {
	var req models.MinioConfig
	if err := c.Bind(&req); err != nil {
		return req, response.RespondError(c, http.StatusBadRequest, "invalid request body")
	}

	// Minio config 검증
	if err := h.minioValidator.ValidateMinioConfig(&req); err != nil {
		return req, echo.NewHTTPError(http.StatusBadRequest, "minio config validation failed: "+err.Error())
	}

	return req, nil
}
