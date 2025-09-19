package minio

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/utils"
)

// Handler : MinIO 관련 HTTP 핸들러
type Handler struct {
	*handler.BaseHandler
	service *Service
}

// NewHandler : 새로운 MinIO 핸들러 생성
func NewHandler(base *handler.BaseHandler) *Handler {
	return &Handler{
		BaseHandler: base,
		service:     NewService(base),
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
	return h.BaseHandler.HealthCheck(c, handler.HealthCheckConfig{
		ServiceName: "minio",
		DefaultNS:   "", // MinIO는 네임스페이스가 없음
		HealthFunc: func(client client.Client, ctx context.Context) error {
			_, err := client.Minio().ListBuckets(ctx)
			return err
		},
	})
}

// ListBuckets : MinIO 버킷 목록 조회
// @Summary List MinIO Buckets
// @Description Get a list of MinIO buckets
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets [post]
func (h *Handler) ListBuckets(c echo.Context) error {
	return h.HandleResourceClient(c, "minio-list-buckets", func(client client.Client, ctx context.Context) (interface{}, error) {
		return h.service.ListBucketsInternal(client, ctx)
	})
}

// CheckBucketExists : 버킷 존재 여부 확인
// @Summary Check Bucket Exists
// @Description Check if a MinIO bucket exists
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket} [get]
func (h *Handler) CheckBucketExists(c echo.Context) error {
	return h.HandleResourceClient(c, "bucket-exists", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름 가져오기
		bucketName := c.Param("bucket")
		if bucketName == "" {
			return nil, echo.NewHTTPError(400, "bucket parameter is required")
		}

		return h.service.CheckBucketExistsInternal(client, ctx, bucketName)
	})
}

// CreateBucket : 버킷 생성
// @Summary Create Bucket
// @Description Create a new MinIO bucket
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket} [post]
func (h *Handler) CreateBucket(c echo.Context) error {
	return h.HandleResourceClient(c, "create-bucket", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름 가져오기
		bucketName := c.Param("bucket")
		if bucketName == "" {
			return nil, echo.NewHTTPError(400, "bucket parameter is required")
		}

		return h.service.CreateBucketInternal(client, ctx, bucketName)
	})
}

// DeleteBucket : 버킷 삭제
// @Summary Delete Bucket
// @Description Delete a MinIO bucket
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket} [delete]
func (h *Handler) DeleteBucket(c echo.Context) error {
	return h.HandleResourceClient(c, "delete-bucket", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름 가져오기
		bucketName := c.Param("bucket")
		if bucketName == "" {
			return nil, echo.NewHTTPError(400, "bucket parameter is required")
		}

		return h.service.DeleteBucketInternal(client, ctx, bucketName)
	})
}

// ListObjects : MinIO 객체 목록 조회
// @Summary List Objects
// @Description Get a list of objects in a MinIO bucket
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket}/objects [post]
func (h *Handler) ListObjects(c echo.Context) error {
	return h.HandleResourceClient(c, "list-objects", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름 가져오기
		bucketName := c.Param("bucket")
		if bucketName == "" {
			return nil, echo.NewHTTPError(400, "bucket parameter is required")
		}

		return h.service.ListObjectsInternal(client, ctx, bucketName)
	})
}

// GetObject : MinIO 객체 다운로드
// @Summary Get Object
// @Description Download an object from MinIO bucket
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Param object path string true "Object name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket}/objects/{object} [get]
func (h *Handler) GetObject(c echo.Context) error {
	return h.HandleResourceClient(c, "get-object", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름과 객체 이름 가져오기
		bucketName := c.Param("bucket")
		objectName := c.Param("*")
		if bucketName == "" || objectName == "" {
			return nil, echo.NewHTTPError(400, "bucket and object parameters are required")
		}

		return h.service.GetObjectInternal(client, ctx, bucketName, objectName)
	})
}

// PutObject : 객체 업로드
// @Summary Put Object
// @Description Upload an object to MinIO bucket
// @Tags minio
// @Accept multipart/form-data
// @Produce json
// @Param bucket path string true "Bucket name"
// @Param object path string true "Object name"
// @Param config formData string true "MinIO configuration JSON"
// @Param file formData file true "File to upload"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket}/objects/{object} [put]
func (h *Handler) PutObject(c echo.Context) error {
	return h.HandleResourceClient(c, "put-object", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름과 객체 이름 가져오기
		bucketName := c.Param("bucket")
		objectName := c.Param("*")
		if bucketName == "" || objectName == "" {
			return nil, echo.NewHTTPError(400, "bucket and object parameters are required")
		}

		return h.service.PutObjectInternal(client, ctx, bucketName, objectName, c)
	})
}

// DeleteObject : MinIO 객체 삭제
// @Summary Delete Object
// @Description Delete an object from MinIO bucket
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Param object path string true "Object name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket}/objects/{object} [delete]
func (h *Handler) DeleteObject(c echo.Context) error {
	return h.HandleResourceClient(c, "delete-object", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름과 객체 이름 가져오기
		bucketName := c.Param("bucket")
		objectName := c.Param("*")
		if bucketName == "" || objectName == "" {
			return nil, echo.NewHTTPError(400, "bucket and object parameters are required")
		}

		return h.service.DeleteObjectInternal(client, ctx, bucketName, objectName)
	})
}

// StatObject : MinIO 객체 정보 조회
// @Summary Stat Object
// @Description Get object information from MinIO bucket
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Param object path string true "Object name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket}/objects/{object}/stat [get]
func (h *Handler) StatObject(c echo.Context) error {
	return h.HandleResourceClient(c, "stat-object", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름과 객체 이름 가져오기
		bucketName := c.Param("bucket")
		objectName := c.Param("*")
		if bucketName == "" || objectName == "" {
			return nil, echo.NewHTTPError(400, "bucket and object parameters are required")
		}

		return h.service.StatObjectInternal(client, ctx, bucketName, objectName)
	})
}

// CopyObject : MinIO 객체 복사
// @Summary Copy Object
// @Description Copy an object within MinIO
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param sourceBucket path string true "Source bucket name"
// @Param sourceObject path string true "Source object name"
// @Param destBucket path string true "Destination bucket name"
// @Param destObject path string true "Destination object name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{sourceBucket}/objects/{sourceObject}/copy/{destBucket}/{destObject} [post]
func (h *Handler) CopyObject(c echo.Context) error {
	return h.HandleResourceClient(c, "copy-object", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 소스와 대상 정보 가져오기
		sourceBucket := c.Param("sourceBucket")
		sourceObject := c.Param("sourceObject")
		destBucket := c.Param("destBucket")
		destObject := c.Param("destObject")
		if sourceBucket == "" || sourceObject == "" || destBucket == "" || destObject == "" {
			return nil, echo.NewHTTPError(400, "source and destination parameters are required")
		}

		return h.service.CopyObjectInternal(client, ctx, sourceBucket, sourceObject, destBucket, destObject)
	})
}

// PresignedGetObject : MinIO 객체 미리 서명된 다운로드 URL 생성
// @Summary Presigned Get Object
// @Description Generate a presigned URL for downloading an object
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Param object path string true "Object name"
// @Param expiry query int false "URL expiry in seconds (default: 3600)"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket}/objects/{object}/presigned-get [get]
func (h *Handler) PresignedGetObject(c echo.Context) error {
	return h.HandleResourceClient(c, "presigned-get-object", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름과 객체 이름 가져오기
		bucketName := c.Param("bucket")
		objectName := c.Param("*")
		if bucketName == "" || objectName == "" {
			return nil, echo.NewHTTPError(400, "bucket and object parameters are required")
		}

		// 만료 시간 가져오기 (기본값: 3600초)
		expiry := utils.ResolveInt(c, "expiry", 3600)

		return h.service.PresignedGetObjectInternal(client, ctx, bucketName, objectName, expiry)
	})
}

// PresignedPutObject : MinIO 객체 미리 서명된 업로드 URL 생성
// @Summary Presigned Put Object
// @Description Generate a presigned URL for uploading an object
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Param object path string true "Object name"
// @Param expiry query int false "URL expiry in seconds (default: 3600)"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket}/objects/{object}/presigned-put [put]
func (h *Handler) PresignedPutObject(c echo.Context) error {
	return h.HandleResourceClient(c, "presigned-put-object", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름과 객체 이름 가져오기
		bucketName := c.Param("bucket")
		objectName := c.Param("object")
		if bucketName == "" || objectName == "" {
			return nil, echo.NewHTTPError(400, "bucket and object parameters are required")
		}

		// 만료 시간 가져오기 (기본값: 3600초)
		expiry := utils.ResolveInt(c, "expiry", 3600)

		return h.service.PresignedPutObjectInternal(client, ctx, bucketName, objectName, expiry)
	})
}
