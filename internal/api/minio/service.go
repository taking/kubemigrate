package minio

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/client/minio"
)

// Service : MinIO 관련 비즈니스 로직
type Service struct {
	*handler.BaseHandler
}

// NewService : 새로운 MinIO 서비스 생성
func NewService(base *handler.BaseHandler) *Service {
	return &Service{
		BaseHandler: base,
	}
}

// GetBucketsInternal : MinIO 버킷 목록 조회 (내부 로직)
func (s *Service) GetBucketsInternal(client client.Client, ctx context.Context) (interface{}, error) {
	// MinIO 버킷 목록 조회
	buckets, err := client.Minio().ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	return buckets, nil
}

// CheckBucketExistsInternal : MinIO 버킷 존재 여부 확인 (내부 로직)
func (s *Service) CheckBucketExistsInternal(client client.Client, ctx context.Context, bucketName string) (interface{}, error) {
	// MinIO 버킷 존재 여부 확인
	exists, err := client.Minio().BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	return map[string]interface{}{
		"bucket":  bucketName,
		"exists":  exists,
		"message": fmt.Sprintf("Bucket '%s' %s", bucketName, map[bool]string{true: "exists", false: "does not exist"}[exists]),
	}, nil
}

// CreateBucketInternal : MinIO 버킷 생성 (내부 로직)
func (s *Service) CreateBucketInternal(client client.Client, ctx context.Context, bucketName string) (interface{}, error) {
	// MinIO 버킷 생성
	err := client.Minio().MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	return map[string]interface{}{
		"bucket":  bucketName,
		"message": fmt.Sprintf("Bucket '%s' created successfully", bucketName),
		"status":  "created",
	}, nil
}

// DeleteBucketInternal : MinIO 버킷 삭제 (내부 로직)
func (s *Service) DeleteBucketInternal(client client.Client, ctx context.Context, bucketName string) (interface{}, error) {
	// MinIO 버킷 삭제
	err := client.Minio().DeleteBucket(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to delete bucket: %w", err)
	}

	return map[string]interface{}{
		"bucket":  bucketName,
		"message": fmt.Sprintf("Bucket '%s' deleted successfully", bucketName),
		"status":  "deleted",
	}, nil
}

// GetObjectsInternal : MinIO 객체 목록 조회 (내부 로직)
func (s *Service) GetObjectsInternal(client client.Client, ctx context.Context, bucketName string) (interface{}, error) {
	// MinIO 객체 목록 조회
	objects, err := client.Minio().ListObjects(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return objects, nil
}

// GetObjectInternal : MinIO 객체 다운로드 (내부 로직)
func (s *Service) GetObjectInternal(client client.Client, ctx context.Context, bucketName, objectName string) (interface{}, error) {
	// MinIO 객체 다운로드
	object, err := client.Minio().GetObject(ctx, bucketName, objectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return map[string]interface{}{
		"bucket":  bucketName,
		"object":  objectName,
		"content": object,
		"status":  "downloaded",
	}, nil
}

// PutObjectInternal : MinIO 객체 업로드 (내부 로직)
func (s *Service) PutObjectInternal(client client.Client, ctx context.Context, bucketName, objectName string, file echo.Context) (interface{}, error) {
	// multipart/form-data에서 파일 처리
	form, err := file.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("multipart form is required: %w", err)
	}

	// 파일 가져오기
	files := form.File["file"]
	if len(files) == 0 {
		return nil, fmt.Errorf("file is required")
	}
	fileHeader := files[0]

	// 파일 열기
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// 객체 업로드
	uploadInfo, err := client.Minio().PutObject(ctx, bucketName, objectName, src, fileHeader.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to upload object: %w", err)
	}

	return map[string]interface{}{
		"bucket":     bucketName,
		"object":     objectName,
		"uploadInfo": uploadInfo,
		"status":     "uploaded",
	}, nil
}

// DeleteObjectInternal : MinIO 객체 삭제 (내부 로직)
func (s *Service) DeleteObjectInternal(client client.Client, ctx context.Context, bucketName, objectName string) (interface{}, error) {
	// MinIO 객체 삭제
	err := client.Minio().DeleteObject(ctx, bucketName, objectName)
	if err != nil {
		return nil, fmt.Errorf("failed to delete object: %w", err)
	}

	return map[string]interface{}{
		"bucket":  bucketName,
		"object":  objectName,
		"message": fmt.Sprintf("Object '%s' deleted successfully from bucket '%s'", objectName, bucketName),
		"status":  "deleted",
	}, nil
}

// StatObjectInternal : MinIO 객체 정보 조회 (내부 로직)
func (s *Service) StatObjectInternal(client client.Client, ctx context.Context, bucketName, objectName string) (interface{}, error) {
	// MinIO 객체 정보 조회
	objectInfo, err := client.Minio().StatObject(ctx, bucketName, objectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	return map[string]interface{}{
		"bucket":     bucketName,
		"object":     objectName,
		"objectInfo": objectInfo,
		"status":     "found",
	}, nil
}

// CopyObjectInternal : MinIO 객체 복사 (내부 로직)
func (s *Service) CopyObjectInternal(client client.Client, ctx context.Context, sourceBucket, sourceObject, destBucket, destObject string) (interface{}, error) {
	// MinIO 객체 복사
	copyInfo, err := client.Minio().CopyObject(ctx, sourceBucket, sourceObject, destBucket, destObject)
	if err != nil {
		return nil, fmt.Errorf("failed to copy object: %w", err)
	}

	return map[string]interface{}{
		"sourceBucket": sourceBucket,
		"sourceObject": sourceObject,
		"destBucket":   destBucket,
		"destObject":   destObject,
		"copyInfo":     copyInfo,
		"status":       "copied",
	}, nil
}

// PresignedGetObjectInternal : MinIO 객체 미리 서명된 다운로드 URL 생성 (내부 로직)
func (s *Service) PresignedGetObjectInternal(client client.Client, ctx context.Context, bucketName, objectName string, expiry int) (interface{}, error) {
	// MinIO 미리 서명된 다운로드 URL 생성
	url, err := client.Minio().PresignedGetObject(ctx, bucketName, objectName, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned get URL: %w", err)
	}

	return map[string]interface{}{
		"bucket": bucketName,
		"object": objectName,
		"url":    url,
		"expiry": expiry,
		"status": "generated",
	}, nil
}

// PresignedPutObjectInternal : MinIO 객체 미리 서명된 업로드 URL 생성 (내부 로직)
func (s *Service) PresignedPutObjectInternal(client client.Client, ctx context.Context, bucketName, objectName string, expiry int) (interface{}, error) {
	// MinIO 미리 서명된 업로드 URL 생성
	url, err := client.Minio().PresignedPutObject(ctx, bucketName, objectName, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned put URL: %w", err)
	}

	return map[string]interface{}{
		"bucket": bucketName,
		"object": objectName,
		"url":    url,
		"expiry": expiry,
		"status": "generated",
	}, nil
}

// DeleteFolderInternal : MinIO에서 폴더와 그 안의 모든 객체를 삭제합니다 (내부 로직)
func (s *Service) DeleteFolderInternal(client client.Client, ctx context.Context, bucketName, folderPath string) (interface{}, error) {
	// 폴더 경로 정규화 (끝에 / 추가)
	if folderPath != "" && folderPath[len(folderPath)-1] != '/' {
		folderPath += "/"
	}

	// MinIO에서 폴더 삭제
	err := client.Minio().DeleteFolder(ctx, bucketName, folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to delete folder %s in bucket %s: %w", folderPath, bucketName, err)
	}

	return map[string]interface{}{
		"bucket":     bucketName,
		"folderPath": folderPath,
		"status":     "deleted",
		"message":    fmt.Sprintf("Successfully deleted folder %s from bucket %s", folderPath, bucketName),
	}, nil
}

// ListObjectsInFolderInternal : 폴더 내 객체 목록을 조회합니다 (내부 로직)
func (s *Service) ListObjectsInFolderInternal(client client.Client, ctx context.Context, bucketName, folderPath string) (interface{}, error) {
	// 폴더 경로 정규화 (끝에 / 추가)
	if folderPath != "" && folderPath[len(folderPath)-1] != '/' {
		folderPath += "/"
	}

	// MinIO에서 폴더 내 객체 목록 조회
	objects, err := client.Minio().ListObjectsInFolder(ctx, bucketName, folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in folder %s in bucket %s: %w", folderPath, bucketName, err)
	}

	// objects를 []interface{}로 변환
	var objectList []interface{}
	if objectsSlice, ok := objects.([]interface{}); ok {
		objectList = objectsSlice
	} else {
		// 다른 타입인 경우 빈 슬라이스로 처리
		objectList = []interface{}{}
	}

	return map[string]interface{}{
		"bucket":     bucketName,
		"folderPath": folderPath,
		"objects":    objects,
		"count":      len(objectList),
	}, nil
}
