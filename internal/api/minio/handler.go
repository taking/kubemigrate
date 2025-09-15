package minio

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	minioSDK "github.com/minio/minio-go/v7"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/config"
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
	return h.HandleResourceClient(c, "minio-health", func(client client.Client, ctx context.Context) (interface{}, error) {
		// MinIO 연결 테스트
		_, err := client.Minio().ListBuckets(ctx)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"service": "minio",
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

		// 버킷 존재 여부 확인
		exists, err := client.Minio().BucketExists(ctx, bucketName)
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

		// 버킷 생성
		err := client.Minio().MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
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
// @Param bucket path string true "Bucket name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/create-if-not-exists [post]
func (h *Handler) CreateBucketIfNotExists(c echo.Context) error {
	return h.HandleResourceClient(c, "create-bucket-if-not-exists", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름 가져오기
		bucketName := c.Param("bucket")
		if bucketName == "" {
			return nil, echo.NewHTTPError(400, "bucket parameter is required")
		}

		// 버킷 존재 여부 확인
		exists, err := client.Minio().BucketExists(ctx, bucketName)
		if err != nil {
			return nil, err
		}

		if !exists {
			// 버킷이 없으면 생성
			err = client.Minio().MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
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

// ListBuckets : 버킷 목록 조회
// @Summary List Buckets
// @Description List all MinIO buckets
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets [get]
func (h *Handler) ListBuckets(c echo.Context) error {
	return h.HandleResourceClient(c, "list-buckets", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 목록 조회
		buckets, err := client.Minio().ListBuckets(ctx)
		if err != nil {
			return nil, err
		}

		// buckets를 []interface{}로 변환
		bucketSlice := buckets.([]minioSDK.BucketInfo)
		bucketList := make([]interface{}, len(bucketSlice))
		for i, bucket := range bucketSlice {
			bucketList[i] = bucket
		}

		return map[string]interface{}{
			"buckets": bucketList,
			"count":   len(bucketSlice),
		}, nil
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

		// 버킷 삭제
		err := client.Minio().DeleteBucket(ctx, bucketName)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"bucket": bucketName,
			"status": "deleted",
		}, nil
	})
}

// ListObjects : 객체 목록 조회
// @Summary List Objects
// @Description List objects in a MinIO bucket
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket}/objects [get]
func (h *Handler) ListObjects(c echo.Context) error {
	return h.HandleResourceClient(c, "list-objects", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름 가져오기
		bucketName := c.Param("bucket")
		if bucketName == "" {
			return nil, echo.NewHTTPError(400, "bucket parameter is required")
		}

		// 객체 목록 조회
		objects, err := client.Minio().ListObjects(ctx, bucketName)
		if err != nil {
			return nil, err
		}

		// objects를 []interface{}로 변환
		objectSlice := objects.([]minioSDK.ObjectInfo)
		objectList := make([]interface{}, len(objectSlice))
		for i, object := range objectSlice {
			objectList[i] = object
		}

		return map[string]interface{}{
			"bucket":  bucketName,
			"objects": objectList,
			"count":   len(objectSlice),
		}, nil
	})
}

// GetObject : 객체 다운로드
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

		// 객체 다운로드
		object, err := client.Minio().GetObject(ctx, bucketName, objectName)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"bucket": bucketName,
			"object": objectName,
			"data":   object,
		}, nil
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
	// 버킷 이름과 객체 이름 가져오기
	bucketName := c.Param("bucket")
	objectName := c.Param("*")
	if bucketName == "" || objectName == "" {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"MISSING_PARAMETERS",
			"Bucket and object parameters are required",
			"Both bucket and object parameters must be provided")
	}

	// multipart/form-data에서 설정과 파일 처리
	form, err := c.MultipartForm()
	if err != nil {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_FORM_DATA",
			"Multipart form is required",
			err.Error())
	}

	// MinIO 설정 가져오기
	configValue := form.Value["config"]
	if len(configValue) == 0 {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"MISSING_CONFIG",
			"Config is required",
			"MinIO configuration must be provided in the form data")
	}

	// JSON 설정 파싱
	var minioConfig config.MinioConfig
	if err := json.Unmarshal([]byte(configValue[0]), &minioConfig); err != nil {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_CONFIG_FORMAT",
			"Invalid config format",
			err.Error())
	}

	// MinIO 설정 검증
	if err := h.MinioValidator.ValidateMinioConfig(&minioConfig); err != nil {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_MINIO_CONFIG",
			"Invalid MinIO configuration",
			err.Error())
	}

	// 파일 가져오기
	files := form.File["file"]
	if len(files) == 0 {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"MISSING_FILE",
			"File is required",
			"At least one file must be provided for upload")
	}
	file := files[0]

	// 파일 열기
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = src.Close() // 에러 무시 (defer에서 에러 반환 불가)
	}()

	// MinIO 클라이언트 생성
	minioClient, err := minio.NewClientWithConfig(minioConfig)
	if err != nil {
		return response.RespondWithErrorModel(c, http.StatusInternalServerError,
			"MINIO_CLIENT_CREATION_FAILED",
			"Failed to create MinIO client",
			err.Error())
	}

	// 객체 업로드
	uploadInfo, err := minioClient.PutObject(c.Request().Context(), bucketName, objectName, src, file.Size)
	if err != nil {
		return response.RespondWithErrorModel(c, http.StatusInternalServerError,
			"OBJECT_UPLOAD_FAILED",
			"Failed to upload object",
			err.Error())
	}

	return response.RespondWithData(c, http.StatusOK, map[string]interface{}{
		"bucket":     bucketName,
		"object":     objectName,
		"uploadInfo": uploadInfo,
		"status":     "uploaded",
	})
}

// DeleteObject : 객체 삭제
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
// @Router /v1/minio/buckets/{bucket}/objects/delete [delete]
func (h *Handler) DeleteObject(c echo.Context) error {
	return h.HandleResourceClient(c, "delete-object", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름과 객체 이름 가져오기
		bucketName := c.Param("bucket")
		objectName := c.Param("*")
		if bucketName == "" || objectName == "" {
			return nil, echo.NewHTTPError(400, "bucket and object parameters are required")
		}

		// 객체 삭제
		err := client.Minio().DeleteObject(ctx, bucketName, objectName)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"bucket": bucketName,
			"object": objectName,
			"status": "deleted",
		}, nil
	})
}

// StatObject : 객체 정보 조회
// @Summary Stat Object
// @Description Get object metadata from MinIO bucket
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Param object path string true "Object name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket}/objects/stat [get]
func (h *Handler) StatObject(c echo.Context) error {
	return h.HandleResourceClient(c, "stat-object", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름과 객체 이름 가져오기
		bucketName := c.Param("bucket")
		objectName := c.Param("*")
		if bucketName == "" || objectName == "" {
			return nil, echo.NewHTTPError(400, "bucket and object parameters are required")
		}

		// 객체 정보 조회
		objectInfo, err := client.Minio().StatObject(ctx, bucketName, objectName)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"bucket":     bucketName,
			"object":     objectName,
			"objectInfo": objectInfo,
		}, nil
	})
}

// CopyObject : 객체 복사
// @Summary Copy Object
// @Description Copy an object within or between MinIO buckets
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param srcBucket path string true "Source bucket name"
// @Param srcObject path string true "Source object name"
// @Param dstBucket path string true "Destination bucket name"
// @Param dstObject path string true "Destination object name"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{srcBucket}/objects/{srcObject}/copy/{dstBucket}/{dstObject} [post]
func (h *Handler) CopyObject(c echo.Context) error {
	return h.HandleResourceClient(c, "copy-object", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 경로에서 소스와 대상 정보 가져오기
		srcBucket := c.Param("srcBucket")
		srcObject := c.Param("srcObject")
		dstBucket := c.Param("dstBucket")
		dstObject := c.Param("dstObject")

		if srcBucket == "" || srcObject == "" || dstBucket == "" || dstObject == "" {
			return nil, echo.NewHTTPError(400, "srcBucket, srcObject, dstBucket, and dstObject parameters are required")
		}

		// 객체 복사
		copyInfo, err := client.Minio().CopyObject(ctx, srcBucket, srcObject, dstBucket, dstObject)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"srcBucket": srcBucket,
			"srcObject": srcObject,
			"dstBucket": dstBucket,
			"dstObject": dstObject,
			"copyInfo":  copyInfo,
			"status":    "copied",
		}, nil
	})
}

// PresignedGetObject : Presigned GET URL 생성
// @Summary Generate Presigned GET URL
// @Description Generate a presigned URL for downloading an object
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Param object path string true "Object name"
// @Param expiry query int false "Expiry time in seconds (default: 3600)"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/minio/buckets/{bucket}/objects/{object}/presigned-get [get]
func (h *Handler) PresignedGetObject(c echo.Context) error {
	return h.HandleResourceClient(c, "presigned-get-object", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 버킷 이름과 객체 이름 가져오기
		bucketName := c.Param("bucket")
		objectName := c.Param("object")
		if bucketName == "" || objectName == "" {
			return nil, echo.NewHTTPError(400, "bucket and object parameters are required")
		}

		// 만료 시간 가져오기 (기본값: 3600초)
		expiry := 3600
		if expiryStr := c.QueryParam("expiry"); expiryStr != "" {
			if parsedExpiry, err := strconv.Atoi(expiryStr); err == nil {
				expiry = parsedExpiry
			}
		}

		// Presigned GET URL 생성
		url, err := client.Minio().PresignedGetObject(ctx, bucketName, objectName, expiry)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"bucket": bucketName,
			"object": objectName,
			"url":    url,
			"expiry": expiry,
		}, nil
	})
}

// PresignedPutObject : Presigned PUT URL 생성
// @Summary Generate Presigned PUT URL
// @Description Generate a presigned URL for uploading an object
// @Tags minio
// @Accept json
// @Produce json
// @Param request body config.MinioConfig true "MinIO configuration"
// @Param bucket path string true "Bucket name"
// @Param object path string true "Object name"
// @Param expiry query int false "Expiry time in seconds (default: 3600)"
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
		expiry := 3600
		if expiryStr := c.QueryParam("expiry"); expiryStr != "" {
			if parsedExpiry, err := strconv.Atoi(expiryStr); err == nil {
				expiry = parsedExpiry
			}
		}

		// Presigned PUT URL 생성
		url, err := client.Minio().PresignedPutObject(ctx, bucketName, objectName, expiry)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"bucket": bucketName,
			"object": objectName,
			"url":    url,
			"expiry": expiry,
		}, nil
	})
}
