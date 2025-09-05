package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/pkg/cache"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/health"
	"github.com/taking/kubemigrate/pkg/interfaces"
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
// @Param request body models.MinioConfigRequest true "Minio connection configuration"
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

// CheckBucketExists : Minio 버킷 존재 여부 확인
// @Summary Minio 버킷 존재 여부 확인
// @Description Check if Minio bucket exists using MinioConfig
// @Tags minio
// @Accept json
// @Produce json
// @Param request body models.MinioConfigRequest true "Minio configuration with bucket name"
// @Success 200 {object} models.SwaggerSuccessResponse "Bucket exists"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /minio/buckets/:name/status [get]
func (h *MinioHandler) CheckBucketExists(c echo.Context) error {
	return h.handleMinioResourceWithCache(c, "bucket-check", func(client interfaces.MinioClient, ctx context.Context) (interface{}, error) {
		bucketName := c.Param("name")

		exists, err := client.BucketExists(ctx, bucketName)
		if err != nil {
			return nil, err
		}

		message := "Bucket does not exist"
		if exists {
			message = "Bucket exists"
		}

		return map[string]interface{}{
			"status":  exists,
			"message": message,
		}, nil
	})
}

// CreateBucket : Minio 버킷 생성
// @Summary Minio 버킷 생성
// @Description MinioConfig를 사용하여 Minio 버킷을 생성합니다
// @Tags minio
// @Accept json
// @Produce json
// @Param request body models.MinioConfigRequest true "Minio configuration with bucket name"
// @Success 200 {object} models.SwaggerSuccessResponse "Bucket created or already exists"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /minio/buckets/:name [post]
func (h *MinioHandler) CreateBucket(c echo.Context) error {
	return h.handleMinioResourceWithCache(c, "bucket-check", func(client interfaces.MinioClient, ctx context.Context) (interface{}, error) {
		bucketName := c.Param("name")

		exists, err := client.BucketExists(ctx, bucketName)
		if err != nil {
			return nil, err
		}

		if !exists {
			if err := client.CreateBucket(ctx, bucketName); err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"status":  "created",
				"message": "Bucket created successfully",
			}, nil
		}

		return map[string]interface{}{
			"status":  "exists",
			"message": "Bucket already exists",
		}, nil
	})
}

// handleMinioResourceWithCache : 캐시를 사용하는 Minio 리소스 처리 헬퍼
func (h *MinioHandler) handleMinioResourceWithCache(c echo.Context, cacheKey string,
	getResource func(interfaces.MinioClient, context.Context) (interface{}, error)) error {

	// MinioConfig 검증
	req, err := utils.BindAndValidateMinioConfig(c, h.minioValidator)
	if err != nil {
		return err
	}

	// 캐시 키 생성
	fullCacheKey := fmt.Sprintf("minio:%s:%s", cacheKey, req.BucketName)

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
		data, err := getResource(minioClient, c.Request().Context())
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
