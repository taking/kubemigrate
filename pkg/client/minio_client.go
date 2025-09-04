package client

import (
	"context"
	"fmt"
	"github.com/taking/kubemigrate/pkg/errors"
	"github.com/taking/kubemigrate/pkg/interfaces"
	"github.com/taking/kubemigrate/pkg/models"
	"github.com/taking/kubemigrate/pkg/utils"
	"github.com/taking/kubemigrate/pkg/validator"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// minioClient : MinIO 클라이언트
type minioClient struct {
	Client *minio.Client
}

// NewMinioClient : MinIO 클라이언트 초기화
func NewMinioClient(cfg models.MinioConfig) (interfaces.MinioClient, error) {

	// Minio Validator
	validator := validator.NewMinioValidator()
	if err := validator.ValidateMinioConfig(&cfg); err != nil {
		return nil, fmt.Errorf("minio config validation failed: %w", err)
	}

	cfg.UseSSL = utils.DefaultBool(cfg.UseSSL, false) // UseSSL이 비어 있으면 기본값 사용

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client (endpoint: %s): %w", cfg.Endpoint, err)
	}

	mc := &minioClient{Client: client}

	// 연결 검증 (5초 제한 타임아웃)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mc.HealthCheck(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to minio (endpoint: %s): %w", cfg.Endpoint, err)
	}

	return mc, nil
}

// HealthCheck : MinIO 서버 연결 확인
func (m *minioClient) HealthCheck(ctx context.Context) error {
	return utils.RunWithTimeout(ctx, func() error {
		_, err := m.Client.ListBuckets(ctx)
		if err != nil {
			return errors.WrapMinioError(err)
		}
		return nil
	})
}

// BucketExists checks if bucket exists
func (m *minioClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return m.Client.BucketExists(ctx, bucketName)
}

// CreateBucket creates a new bucket
func (m *minioClient) CreateBucket(ctx context.Context, bucketName string) error {
	return m.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

// CreateBucketIfNotExists : 버킷 확인 및 생성
func (m *minioClient) CreateBucketIfNotExists(ctx context.Context, bucketName, region string) (string, error) {
	// 버킷 존재 여부 확인
	exists, err := m.BucketExists(ctx, bucketName)
	if err != nil {
		return "", fmt.Errorf("failed to check bucket existence: %w", err)
	}

	// 버킷이 존재할 경우
	if exists {
		return fmt.Sprintf("Bucket '%s' already exists", bucketName), nil
	}

	// 버킷이 존재하지 않은 경우 생성 시도
	err = m.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: region})
	if err != nil {
		return "", fmt.Errorf("failed to create bucket '%s': %w", bucketName, err)
	}

	return fmt.Sprintf("Bucket '%s' created successfully", bucketName), nil
}
