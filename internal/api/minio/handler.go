package minio

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/utils"
)

// Handler : MinIO 관련 HTTP 핸들러
type Handler struct {
	*handler.BaseHandler
}

// NewHandler : 새로운 MinIO 핸들러 생성
func NewHandler(base *handler.BaseHandler) *Handler {
	return &Handler{
		BaseHandler: base,
	}
}

// HealthCheck : MinIO 연결 상태 확인
// @Summary MinIO Connection Test
// @Description Test MinIO connection with provided configuration
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/health [post]
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.HandleMinioResource(c, "minio-health", func(minioClient minio.Client, ctx context.Context) (interface{}, error) {
		// MinIO 연결 테스트
		_, err := minioClient.ListBuckets(ctx)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"service": "minio",
			"status":  "healthy",
			"message": "MinIO connection is working",
		}, nil
	})
}

// CheckBucketExists : 버킷 존재 여부 확인
// @Summary Check Bucket Exists
// @Description Check if a MinIO bucket exists
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket query string true "Bucket name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/bucket/exists [post]
func (h *Handler) CheckBucketExists(c echo.Context) error {
	return h.HandleMinioResource(c, "bucket-exists", func(minioClient minio.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		_, err := utils.BindAndValidateMinioConfig(c, h.MinioValidator)
		if err != nil {
			return nil, err
		}

		// 버킷 이름 가져오기
		bucketName := c.QueryParam("bucket")
		if bucketName == "" {
			return nil, echo.NewHTTPError(400, "bucket parameter is required")
		}

		// 버킷 존재 여부 확인
		exists, err := minioClient.BucketExists(ctx, bucketName)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"bucket": bucketName,
			"exists": exists,
		}, nil
	})
}

// CreateBucket : 버킷 생성
// @Summary Create Bucket
// @Description Create a new MinIO bucket
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket query string true "Bucket name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/bucket/create [post]
func (h *Handler) CreateBucket(c echo.Context) error {
	return h.HandleMinioResource(c, "create-bucket", func(minioClient minio.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		_, err := utils.BindAndValidateMinioConfig(c, h.MinioValidator)
		if err != nil {
			return nil, err
		}

		// 버킷 이름 가져오기
		bucketName := c.QueryParam("bucket")
		if bucketName == "" {
			return nil, echo.NewHTTPError(400, "bucket parameter is required")
		}

		// 버킷 생성
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"bucket": bucketName,
			"status": "created",
		}, nil
	})
}

// CreateBucketIfNotExists : 버킷이 없으면 생성
// @Summary Create Bucket If Not Exists
// @Description Create a MinIO bucket if it doesn't exist
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket query string true "Bucket name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/bucket/create-if-not-exists [post]
func (h *Handler) CreateBucketIfNotExists(c echo.Context) error {
	return h.HandleMinioResource(c, "create-bucket-if-not-exists", func(minioClient minio.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		_, err := utils.BindAndValidateMinioConfig(c, h.MinioValidator)
		if err != nil {
			return nil, err
		}

		// 버킷 이름 가져오기
		bucketName := c.QueryParam("bucket")
		if bucketName == "" {
			return nil, echo.NewHTTPError(400, "bucket parameter is required")
		}

		// 버킷 존재 여부 확인
		exists, err := minioClient.BucketExists(ctx, bucketName)
		if err != nil {
			return nil, err
		}

		if !exists {
			// 버킷이 없으면 생성
			err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			if err != nil {
				return nil, err
			}
		}

		return map[string]interface{}{
			"bucket": bucketName,
			"exists": true,
			"status": "created_or_exists",
		}, nil
	})
}
