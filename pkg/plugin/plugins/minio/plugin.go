package minio

import (
	"context"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/errors"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/internal/validator"
	"github.com/taking/kubemigrate/pkg/client/minio"
)

// MinioPlugin MinIO 플러그인
type MinioPlugin struct {
	client    minio.Client
	config    map[string]interface{}
	validator *validator.MinioValidator
}

// NewPlugin 새로운 MinIO 플러그인 생성
func NewPlugin() *MinioPlugin {
	return &MinioPlugin{
		validator: validator.NewMinioValidator(),
	}
}

// Name 플러그인 이름
func (p *MinioPlugin) Name() string {
	return "minio"
}

// Version 플러그인 버전
func (p *MinioPlugin) Version() string {
	return "1.0.0"
}

// Description 플러그인 설명
func (p *MinioPlugin) Description() string {
	return "MinIO object storage management plugin"
}

// Initialize 플러그인 초기화
func (p *MinioPlugin) Initialize(config map[string]interface{}) error {
	p.config = config

	// MinIO 클라이언트 초기화
	p.client = minio.NewClient()

	return nil
}

// Shutdown 플러그인 종료
func (p *MinioPlugin) Shutdown() error {
	// 정리 작업이 필요한 경우 여기에 구현
	return nil
}

// RegisterRoutes 라우트 등록
func (p *MinioPlugin) RegisterRoutes(router *echo.Group) error {
	// MinIO 관련 라우트 등록
	minioGroup := router.Group("/minio")

	// 헬스체크
	minioGroup.POST("/health", p.HealthCheckHandler)

	// 버킷 관리 (기존과 동일)
	minioGroup.GET("/buckets/:bucket", p.CheckBucketExistsHandler) // 버킷 존재 확인
	minioGroup.POST("/buckets/:bucket", p.CreateBucketHandler)     // 버킷 생성
	minioGroup.GET("/buckets", p.ListBucketsHandler)               // 버킷 목록 조회
	minioGroup.DELETE("/buckets/:bucket", p.DeleteBucketHandler)   // 버킷 삭제

	// 객체 관리 (기존과 동일)
	minioGroup.GET("/buckets/:bucket/objects", p.ListObjectsHandler)                                          // 객체 목록 조회
	minioGroup.POST("/buckets/:bucket/objects/*", p.PutObjectHandler)                                         // 객체 업로드
	minioGroup.GET("/buckets/:bucket/objects/*", p.GetObjectHandler)                                          // 객체 다운로드
	minioGroup.GET("/buckets/:bucket/objects/*", p.StatObjectHandler)                                         // 객체 정보 조회
	minioGroup.POST("/buckets/:srcBucket/objects/:srcObject/copy/:dstBucket/:dstObject", p.CopyObjectHandler) // 객체 복사
	minioGroup.DELETE("/buckets/:bucket/objects/*", p.DeleteObjectHandler)                                    // 객체 삭제

	// Presigned URL (기존과 동일)
	minioGroup.GET("/buckets/:bucket/objects/:object/presigned-get", p.PresignedGetObjectHandler) // Presigned GET URL 생성
	minioGroup.PUT("/buckets/:bucket/objects/:object/presigned-put", p.PresignedPutObjectHandler) // Presigned PUT URL 생성

	return nil
}

// HealthCheck 헬스체크
func (p *MinioPlugin) HealthCheck(ctx context.Context) error {
	_, err := p.client.ListBuckets(ctx)
	return err
}

// GetServiceType 서비스 타입
func (p *MinioPlugin) GetServiceType() string {
	return "minio"
}

// GetClient 클라이언트 반환
func (p *MinioPlugin) GetClient() interface{} {
	return p.client
}

// SetPluginManager 플러그인 매니저 설정
func (p *MinioPlugin) SetPluginManager(manager interface{}) {
	// Minio 플러그인에서는 현재 사용하지 않음
}

// HealthCheckHandler 헬스체크 핸들러
func (p *MinioPlugin) HealthCheckHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	// MinIO 연결 테스트
	_, err := p.client.ListBuckets(c.Request().Context())
	if err != nil {
		return errors.NewExternalError("minio", "ListBuckets", err)
	}

	return response.RespondWithSuccessModel(c, 200, "MinIO connection is working", map[string]interface{}{
		"service": "minio",
		"message": "MinIO connection is working",
	})
}

// CheckBucketExistsHandler 버킷 존재 여부 확인 핸들러
func (p *MinioPlugin) CheckBucketExistsHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	bucketName := c.Param("bucket")
	if bucketName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing bucket name", "bucket parameter is required")
	}

	exists, err := p.client.BucketExists(c.Request().Context(), bucketName)
	if err != nil {
		return errors.NewExternalError("minio", "BucketExists", err)
	}

	return response.RespondWithData(c, 200, map[string]interface{}{
		"bucket": bucketName,
		"exists": exists,
	})
}

// CreateBucketHandler 버킷 생성 핸들러
func (p *MinioPlugin) CreateBucketHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	bucketName := c.Param("bucket")
	if bucketName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing bucket name", "bucket parameter is required")
	}

	// 버킷이 이미 존재하는지 확인
	exists, err := p.client.BucketExists(c.Request().Context(), bucketName)
	if err != nil {
		return errors.NewExternalError("minio", "BucketExists", err)
	}

	if exists {
		return response.RespondWithMessage(c, 200, "Bucket already exists")
	}

	// 버킷 생성
	err = p.client.MakeBucket(c.Request().Context(), bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return errors.NewExternalError("minio", "MakeBucket", err)
	}

	return response.RespondWithMessage(c, 201, "Bucket created successfully")
}

// ListBucketsHandler 버킷 목록 조회 핸들러
func (p *MinioPlugin) ListBucketsHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	buckets, err := p.client.ListBuckets(c.Request().Context())
	if err != nil {
		return errors.NewExternalError("minio", "ListBuckets", err)
	}

	return response.RespondWithData(c, 200, buckets)
}

// DeleteBucketHandler 버킷 삭제 핸들러
func (p *MinioPlugin) DeleteBucketHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	bucketName := c.Param("bucket")
	if bucketName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing bucket name", "bucket parameter is required")
	}

	err := p.client.DeleteBucket(c.Request().Context(), bucketName)
	if err != nil {
		return errors.NewExternalError("minio", "DeleteBucket", err)
	}

	return response.RespondWithMessage(c, 200, "Bucket deleted successfully")
}

// ListObjectsHandler 객체 목록 조회 핸들러
func (p *MinioPlugin) ListObjectsHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	bucketName := c.Param("bucket")
	if bucketName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing bucket name", "bucket parameter is required")
	}

	objects, err := p.client.ListObjects(c.Request().Context(), bucketName)
	if err != nil {
		return errors.NewExternalError("minio", "ListObjects", err)
	}

	return response.RespondWithData(c, 200, objects)
}

// GetObjectHandler 객체 조회 핸들러
func (p *MinioPlugin) GetObjectHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	bucketName := c.Param("bucket")
	objectName := c.Param("*")
	if bucketName == "" || objectName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing required parameters", "bucket and object parameters are required")
	}

	object, err := p.client.GetObject(c.Request().Context(), bucketName, objectName)
	if err != nil {
		return errors.NewExternalError("minio", "GetObject", err)
	}

	return response.RespondWithData(c, 200, object)
}

// DeleteObjectHandler 객체 삭제 핸들러
func (p *MinioPlugin) DeleteObjectHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	bucketName := c.Param("bucket")
	objectName := c.Param("*")
	if bucketName == "" || objectName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing required parameters", "bucket and object parameters are required")
	}

	err := p.client.DeleteObject(c.Request().Context(), bucketName, objectName)
	if err != nil {
		return errors.NewExternalError("minio", "DeleteObject", err)
	}

	return response.RespondWithMessage(c, 200, "Object deleted successfully")
}

// StatObjectHandler 객체 정보 조회 핸들러
func (p *MinioPlugin) StatObjectHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	bucketName := c.Param("bucket")
	objectName := c.Param("*")
	if bucketName == "" || objectName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing required parameters", "bucket and object parameters are required")
	}

	objectInfo, err := p.client.StatObject(c.Request().Context(), bucketName, objectName)
	if err != nil {
		return errors.NewExternalError("minio", "StatObject", err)
	}

	return response.RespondWithData(c, 200, objectInfo)
}

// CopyObjectHandler 객체 복사 핸들러
func (p *MinioPlugin) CopyObjectHandler(c echo.Context) error {
	var req struct {
		config.MinioConfig
		SourceBucket string `json:"sourceBucket"`
		SourceObject string `json:"sourceObject"`
		DestBucket   string `json:"destBucket"`
		DestObject   string `json:"destObject"`
	}

	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	bucketName := c.Param("bucket")
	objectName := c.Param("*")
	if bucketName == "" || objectName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing required parameters", "bucket and object parameters are required")
	}

	copyInfo, err := p.client.CopyObject(c.Request().Context(), req.SourceBucket, req.SourceObject, req.DestBucket, req.DestObject)
	if err != nil {
		return errors.NewExternalError("minio", "CopyObject", err)
	}

	return response.RespondWithData(c, 200, copyInfo)
}

// PresignedGetObjectHandler Presigned GET URL 생성 핸들러
func (p *MinioPlugin) PresignedGetObjectHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	bucketName := c.Param("bucket")
	objectName := c.Param("*")
	if bucketName == "" || objectName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing required parameters", "bucket and object parameters are required")
	}

	// 만료 시간 파라미터 처리
	expiryStr := c.QueryParam("expiry")
	expiry := 3600 // 기본값 1시간
	if expiryStr != "" {
		if parsedExpiry, err := strconv.Atoi(expiryStr); err == nil {
			expiry = parsedExpiry
		}
	}

	url, err := p.client.PresignedGetObject(c.Request().Context(), bucketName, objectName, expiry)
	if err != nil {
		return errors.NewExternalError("minio", "PresignedGetObject", err)
	}

	return response.RespondWithData(c, 200, map[string]interface{}{
		"bucket": bucketName,
		"object": objectName,
		"expiry": expiry,
		"url":    url,
	})
}

// PresignedPutObjectHandler Presigned PUT URL 생성 핸들러
func (p *MinioPlugin) PresignedPutObjectHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	bucketName := c.Param("bucket")
	objectName := c.Param("*")
	if bucketName == "" || objectName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing required parameters", "bucket and object parameters are required")
	}

	// 만료 시간 파라미터 처리
	expiryStr := c.QueryParam("expiry")
	expiry := 3600 // 기본값 1시간
	if expiryStr != "" {
		if parsedExpiry, err := strconv.Atoi(expiryStr); err == nil {
			expiry = parsedExpiry
		}
	}

	url, err := p.client.PresignedPutObject(c.Request().Context(), bucketName, objectName, expiry)
	if err != nil {
		return errors.NewExternalError("minio", "PresignedPutObject", err)
	}

	return response.RespondWithData(c, 200, map[string]interface{}{
		"bucket": bucketName,
		"object": objectName,
		"expiry": expiry,
		"url":    url,
	})
}

// PutObjectHandler 객체 업로드 핸들러
func (p *MinioPlugin) PutObjectHandler(c echo.Context) error {
	var req config.MinioConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	bucketName := c.Param("bucket")
	objectName := c.Param("*")
	if bucketName == "" || objectName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing required parameters", "bucket and object parameters are required")
	}

	// 파일 업로드 처리
	file, err := c.FormFile("file")
	if err != nil {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing file", "file parameter is required")
	}

	src, err := file.Open()
	if err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid file", err.Error())
	}
	defer func() {
		_ = src.Close() // 파일 업로드가 이미 완료된 경우 에러는 무시
	}()

	_, err = p.client.PutObject(c.Request().Context(), bucketName, objectName, src, file.Size)
	if err != nil {
		return errors.NewExternalError("minio", "PutObject", err)
	}

	return response.RespondWithMessage(c, 200, "Object uploaded successfully")
}
