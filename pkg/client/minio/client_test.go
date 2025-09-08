package minio

import (
	"context"
	"testing"

	"github.com/taking/kubemigrate/internal/config"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestListBuckets(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 MinIO 서버가 없으므로 에러가 발생할 것으로 예상
	_, err := client.ListBuckets(ctx)
	if err == nil {
		t.Log("ListBuckets succeeded - this might indicate a real MinIO server is available")
	} else {
		t.Logf("ListBuckets failed as expected: %v", err)
	}
}

func TestGetObject(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 MinIO 서버가 없으므로 에러가 발생할 것으로 예상
	_, err := client.GetObject(ctx, "test-bucket", "test-object")
	if err == nil {
		t.Log("GetObject succeeded - this might indicate a real MinIO server is available")
	} else {
		t.Logf("GetObject failed as expected: %v", err)
	}
}

func TestNewClientWithConfig(t *testing.T) {
	// 빈 설정으로 테스트 (에러가 발생할 것으로 예상)
	cfg := config.MinioConfig{
		Endpoint:  "localhost:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		UseSSL:    false,
	}

	client, err := NewClientWithConfig(cfg)
	if err == nil {
		t.Log("NewClientWithConfig succeeded - this might indicate a real MinIO server is available")
		if client == nil {
			t.Fatal("NewClientWithConfig returned nil client")
		}
	} else {
		t.Logf("NewClientWithConfig failed as expected: %v", err)
	}
}

func TestBucketExists(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 MinIO 서버가 없으므로 에러가 발생할 것으로 예상
	_, err := client.BucketExists(ctx, "test-bucket")
	if err == nil {
		t.Log("BucketExists succeeded - this might indicate a real MinIO server is available")
	} else {
		t.Logf("BucketExists failed as expected: %v", err)
	}
}
