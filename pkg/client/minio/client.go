package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/taking/kubemigrate/pkg/config"
)

// MakeBucketOptions 버킷 생성 옵션
type MakeBucketOptions struct {
	Region string
}

// Client MinIO 클라이언트 인터페이스
type Client interface {
	// Bucket 관련
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	CreateBucket(ctx context.Context, bucketName string) error
	MakeBucket(ctx context.Context, bucketName string, opts MakeBucketOptions) error
	CreateBucketIfNotExists(ctx context.Context, bucketName string) error
	DeleteBucket(ctx context.Context, bucketName string) error
	ListBuckets(ctx context.Context) (interface{}, error)

	// Object 관련
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64) (interface{}, error)
	GetObject(ctx context.Context, bucketName, objectName string) (interface{}, error)
	DeleteObject(ctx context.Context, bucketName, objectName string) error
	ListObjects(ctx context.Context, bucketName string) (interface{}, error)

	// Object 정보
	StatObject(ctx context.Context, bucketName, objectName string) (interface{}, error)
	CopyObject(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string) (interface{}, error)

	// Presigned URL
	PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry int) (string, error)
	PresignedPutObject(ctx context.Context, bucketName, objectName string, expiry int) (string, error)

	// HealthCheck : MinIO 연결 확인
	HealthCheck(ctx context.Context) error
}

// client MinIO 클라이언트 구현체
type client struct {
	minioClient *minio.Client
}

// NewClient : 새로운 MinIO 클라이언트를 생성합니다 (기본 설정)
func NewClient() (Client, error) {
	// 기본 설정으로 더미 클라이언트 생성
	return &client{}, nil
}

// NewClientWithConfig 설정을 받아서 MinIO 클라이언트를 생성합니다
func NewClientWithConfig(cfg config.MinioConfig) (Client, error) {
	// 설정 검증
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("minio endpoint is required")
	}
	if cfg.AccessKey == "" {
		return nil, fmt.Errorf("minio access key is required")
	}
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("minio secret key is required")
	}

	// MinIO 클라이언트 초기화
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client with endpoint '%s': %w", cfg.Endpoint, err)
	}

	return &client{
		minioClient: minioClient,
	}, nil
}

// BucketExists 버킷이 존재하는지 확인합니다
func (c *client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	if c.minioClient == nil {
		return false, fmt.Errorf("minio client not initialized")
	}
	return c.minioClient.BucketExists(ctx, bucketName)
}

// CreateBucket 버킷을 생성합니다
func (c *client) CreateBucket(ctx context.Context, bucketName string) error {
	if c.minioClient == nil {
		return fmt.Errorf("minio client not initialized")
	}
	return c.minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

// MakeBucket 버킷을 생성합니다 (옵션 포함)
func (c *client) MakeBucket(ctx context.Context, bucketName string, opts MakeBucketOptions) error {
	if c.minioClient == nil {
		return fmt.Errorf("minio client not initialized")
	}
	return c.minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region: opts.Region,
	})
}

// CreateBucketIfNotExists 버킷이 없으면 생성합니다
func (c *client) CreateBucketIfNotExists(ctx context.Context, bucketName string) error {
	if c.minioClient == nil {
		return fmt.Errorf("minio client not initialized")
	}

	exists, err := c.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		return c.minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	}

	return nil
}

// DeleteBucket 버킷을 삭제합니다
func (c *client) DeleteBucket(ctx context.Context, bucketName string) error {
	if c.minioClient == nil {
		return fmt.Errorf("minio client not initialized")
	}
	return c.minioClient.RemoveBucket(ctx, bucketName)
}

// ListBuckets 버킷 목록을 조회합니다
func (c *client) ListBuckets(ctx context.Context) (interface{}, error) {
	if c.minioClient == nil {
		return nil, fmt.Errorf("minio client not initialized")
	}
	return c.minioClient.ListBuckets(ctx)
}

// PutObject 객체를 업로드합니다
func (c *client) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64) (interface{}, error) {
	if c.minioClient == nil {
		return nil, fmt.Errorf("minio client not initialized")
	}

	uploadInfo, err := c.minioClient.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}

	return uploadInfo, nil
}

// GetObject 객체를 다운로드합니다
func (c *client) GetObject(ctx context.Context, bucketName, objectName string) (interface{}, error) {
	if c.minioClient == nil {
		return nil, fmt.Errorf("minio client not initialized")
	}

	object, err := c.minioClient.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return object, nil
}

// DeleteObject 객체를 삭제합니다
func (c *client) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	if c.minioClient == nil {
		return fmt.Errorf("minio client not initialized")
	}

	return c.minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

// ListObjects 객체 목록을 조회합니다
func (c *client) ListObjects(ctx context.Context, bucketName string) (interface{}, error) {
	if c.minioClient == nil {
		return nil, fmt.Errorf("minio client not initialized")
	}

	var objects []minio.ObjectInfo
	objectCh := c.minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// StatObject 객체 정보를 조회합니다
func (c *client) StatObject(ctx context.Context, bucketName, objectName string) (interface{}, error) {
	if c.minioClient == nil {
		return nil, fmt.Errorf("minio client not initialized")
	}

	objectInfo, err := c.minioClient.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	return objectInfo, nil
}

// CopyObject 객체를 복사합니다
func (c *client) CopyObject(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string) (interface{}, error) {
	if c.minioClient == nil {
		return nil, fmt.Errorf("minio client not initialized")
	}

	// 소스 객체 정보
	src := minio.CopySrcOptions{
		Bucket: srcBucket,
		Object: srcObject,
	}

	// 대상 객체 설정
	dst := minio.CopyDestOptions{
		Bucket: dstBucket,
		Object: dstObject,
	}

	uploadInfo, err := c.minioClient.CopyObject(ctx, dst, src)
	if err != nil {
		return nil, err
	}

	return uploadInfo, nil
}

// PresignedGetObject Presigned GET URL을 생성합니다
func (c *client) PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry int) (string, error) {
	if c.minioClient == nil {
		return "", fmt.Errorf("minio client not initialized")
	}

	url, err := c.minioClient.PresignedGetObject(ctx, bucketName, objectName, time.Duration(expiry)*time.Second, nil)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

// PresignedPutObject Presigned PUT URL을 생성합니다
func (c *client) PresignedPutObject(ctx context.Context, bucketName, objectName string, expiry int) (string, error) {
	if c.minioClient == nil {
		return "", fmt.Errorf("minio client not initialized")
	}

	url, err := c.minioClient.PresignedPutObject(ctx, bucketName, objectName, time.Duration(expiry)*time.Second)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

// HealthCheck : MinIO 연결 확인
func (c *client) HealthCheck(ctx context.Context) error {
	if c.minioClient == nil {
		return fmt.Errorf("minio client not initialized")
	}

	// 간단한 API 호출로 연결 상태 확인
	_, err := c.minioClient.ListBuckets(ctx)
	return err
}
