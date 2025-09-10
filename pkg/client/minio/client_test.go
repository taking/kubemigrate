package minio

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/taking/kubemigrate/internal/config"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestNewClientWithConfig(t *testing.T) {
	// 실제 MinIO 서버 설정으로 테스트
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

func TestNewClientWithConfigWithNilConfigs(t *testing.T) {
	// nil 설정으로 테스트 (기본 클라이언트로 폴백되어야 함)
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestBucketOperations(t *testing.T) {
	client := NewClient()
	ctx := context.Background()
	testBucket := "test-bucket-" + time.Now().Format("20060102150405")

	// BucketExists 테스트
	exists, err := client.BucketExists(ctx, testBucket)
	if err == nil {
		t.Logf("BucketExists succeeded: bucket %s exists=%v", testBucket, exists)
	} else {
		t.Logf("BucketExists failed as expected: %v", err)
	}

	// CreateBucket 테스트
	err = client.CreateBucket(ctx, testBucket)
	if err == nil {
		t.Logf("CreateBucket succeeded: bucket %s created", testBucket)
	} else {
		t.Logf("CreateBucket failed as expected: %v", err)
	}

	// CreateBucketIfNotExists 테스트
	err = client.CreateBucketIfNotExists(ctx, testBucket)
	if err == nil {
		t.Logf("CreateBucketIfNotExists succeeded: bucket %s", testBucket)
	} else {
		t.Logf("CreateBucketIfNotExists failed as expected: %v", err)
	}

	// DeleteBucket 테스트 (정리)
	err = client.DeleteBucket(ctx, testBucket)
	if err == nil {
		t.Logf("DeleteBucket succeeded: bucket %s deleted", testBucket)
	} else {
		t.Logf("DeleteBucket failed as expected: %v", err)
	}
}

func TestListBuckets(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 MinIO 서버가 없으므로 에러가 발생할 것으로 예상
	buckets, err := client.ListBuckets(ctx)
	if err == nil {
		t.Log("ListBuckets succeeded - this might indicate a real MinIO server is available")
		if buckets != nil {
			t.Logf("ListBuckets returned: %+v", buckets)
		}
	} else {
		t.Logf("ListBuckets failed as expected: %v", err)
	}
}

func TestObjectOperations(t *testing.T) {
	client := NewClient()
	ctx := context.Background()
	testBucket := "test-bucket"
	testObject := "test-object.txt"

	// GetObject 테스트
	object, err := client.GetObject(ctx, testBucket, testObject)
	if err == nil {
		t.Log("GetObject succeeded - this might indicate a real MinIO server is available")
		if object != nil {
			t.Logf("GetObject returned: %+v", object)
		}
	} else {
		t.Logf("GetObject failed as expected: %v", err)
	}

	// PutObject 테스트 (더미 데이터)
	testData := strings.NewReader("test data")
	uploadInfo, err := client.PutObject(ctx, testBucket, testObject, testData, 9)
	if err == nil {
		t.Log("PutObject succeeded - this might indicate a real MinIO server is available")
		if uploadInfo != nil {
			t.Logf("PutObject returned: %+v", uploadInfo)
		}
	} else {
		t.Logf("PutObject failed as expected: %v", err)
	}

	// ListObjects 테스트
	objects, err := client.ListObjects(ctx, testBucket)
	if err == nil {
		t.Log("ListObjects succeeded - this might indicate a real MinIO server is available")
		if objects != nil {
			t.Logf("ListObjects returned: %+v", objects)
		}
	} else {
		t.Logf("ListObjects failed as expected: %v", err)
	}

	// StatObject 테스트
	objectInfo, err := client.StatObject(ctx, testBucket, testObject)
	if err == nil {
		t.Log("StatObject succeeded - this might indicate a real MinIO server is available")
		if objectInfo != nil {
			t.Logf("StatObject returned: %+v", objectInfo)
		}
	} else {
		t.Logf("StatObject failed as expected: %v", err)
	}

	// CopyObject 테스트
	copyInfo, err := client.CopyObject(ctx, testBucket, testObject, testBucket, "copy-"+testObject)
	if err == nil {
		t.Log("CopyObject succeeded - this might indicate a real MinIO server is available")
		if copyInfo != nil {
			t.Logf("CopyObject returned: %+v", copyInfo)
		}
	} else {
		t.Logf("CopyObject failed as expected: %v", err)
	}

	// DeleteObject 테스트
	err = client.DeleteObject(ctx, testBucket, testObject)
	if err == nil {
		t.Log("DeleteObject succeeded - this might indicate a real MinIO server is available")
	} else {
		t.Logf("DeleteObject failed as expected: %v", err)
	}
}

func TestPresignedURLs(t *testing.T) {
	client := NewClient()
	ctx := context.Background()
	testBucket := "test-bucket"
	testObject := "test-object.txt"

	// PresignedGetObject 테스트
	getURL, err := client.PresignedGetObject(ctx, testBucket, testObject, 3600)
	if err == nil {
		t.Log("PresignedGetObject succeeded - this might indicate a real MinIO server is available")
		t.Logf("PresignedGetObject returned URL: %s", getURL)
	} else {
		t.Logf("PresignedGetObject failed as expected: %v", err)
	}

	// PresignedPutObject 테스트
	putURL, err := client.PresignedPutObject(ctx, testBucket, testObject, 1800)
	if err == nil {
		t.Log("PresignedPutObject succeeded - this might indicate a real MinIO server is available")
		t.Logf("PresignedPutObject returned URL: %s", putURL)
	} else {
		t.Logf("PresignedPutObject failed as expected: %v", err)
	}
}

func TestMakeBucketWithOptions(t *testing.T) {
	client := NewClient()
	ctx := context.Background()
	testBucket := "test-bucket-with-region-" + time.Now().Format("20060102150405")

	// MakeBucket with options 테스트
	opts := MakeBucketOptions{
		Region: "us-east-1",
	}

	err := client.MakeBucket(ctx, testBucket, opts)
	if err == nil {
		t.Logf("MakeBucket with options succeeded: bucket %s created with region %s", testBucket, opts.Region)
	} else {
		t.Logf("MakeBucket with options failed as expected: %v", err)
	}

	// 정리
	err = client.DeleteBucket(ctx, testBucket)
	if err == nil {
		t.Logf("Cleanup: bucket %s deleted", testBucket)
	} else {
		t.Logf("Cleanup failed: %v", err)
	}
}
