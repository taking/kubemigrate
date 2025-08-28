package clients

import (
	"context"
	"fmt"
	"taking.kr/velero/models"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Client *minio.Client
}

// NewMinioClient : MinIO 클라이언트 초기화
func NewMinioClient(cfg models.MinioConfig) (*MinioClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}
	return &MinioClient{Client: client}, nil
}

// HealthCheck : MinIO 서버 연결 확인
func (m *MinioClient) HealthCheck(ctx context.Context) error {
	// 5초 제한 타임아웃
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 서버 연결 확인 (버킷 목록 조회 시도)
	_, err := m.Client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("minio health check failed: %w", err)
	}
	return nil
}

// CreateBucketIfNotExists : 버킷 확인 및 생성
func (m *MinioClient) CreateBucketIfNotExists(ctx context.Context, bucketName, region string) (string, error) {
	// 버킷 존재 여부 확인
	exists, err := m.Client.BucketExists(ctx, bucketName)
	if err != nil {
		return "", fmt.Errorf("failed to check bucket existence: %w", err)
	}

	// 버킷이 존재할 경우
	if exists {
		return fmt.Sprintf("✅ Bucket '%s' already exists", bucketName), nil
	}

	// 버킷이 존재하지 않은 경우 생성 시도
	err = m.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: region})
	if err != nil {
		return "", fmt.Errorf("❌ failed to create bucket '%s': %w", bucketName, err)
	}

	return fmt.Sprintf("🎉 Bucket '%s' created successfully", bucketName), nil
}
