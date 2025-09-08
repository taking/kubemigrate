package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taking/kubemigrate/pkg/client/minio"
)

func main() {
	fmt.Println("=== MinIO Client Usage Example ===")

	// 1. MinIO 클라이언트 생성
	fmt.Println("\n1. Creating MinIO client...")
	client := minio.NewClient()
	ctx := context.Background()

	// 클라이언트 생성 후 간단한 테스트로 연결 확인
	_, err := client.ListBuckets(ctx)
	if err != nil {
		log.Printf("MinIO client connection failed: %v", err)
		fmt.Println("Cannot connect to MinIO server. Please check your configuration.")
		fmt.Println("Environment variables to check:")
		fmt.Println("  - MINIO_ENDPOINT (e.g., localhost:9000)")
		fmt.Println("  - MINIO_ACCESS_KEY (e.g., minioadmin)")
		fmt.Println("  - MINIO_SECRET_KEY (e.g., minioadmin)")
		return
	}
	fmt.Println("✅ MinIO client created successfully.")

	// 2. 버킷 목록 조회
	fmt.Println("\n2. Retrieving bucket list...")
	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		log.Printf("Failed to retrieve bucket list: %v", err)
	} else {
		fmt.Printf("✅ Bucket list retrieved successfully (type: %T)\n", buckets)
		// buckets는 interface{} 타입이므로 타입 어설션 필요
		if bucketList, ok := buckets.([]interface{}); ok {
			fmt.Printf("  Found %d buckets in total\n", len(bucketList))
		} else {
			fmt.Printf("  Bucket list type: %T\n", buckets)
		}
	}

	// 3. 테스트 버킷 존재 여부 확인
	testBucketName := "test-bucket-example"
	fmt.Printf("\n3. Checking if test bucket '%s' exists...\n", testBucketName)

	exists, err := client.BucketExists(ctx, testBucketName)
	if err != nil {
		log.Printf("Failed to check bucket existence: %v", err)
	} else {
		if exists {
			fmt.Printf("✅ Bucket '%s' already exists.\n", testBucketName)
		} else {
			fmt.Printf("ℹ️  Bucket '%s' does not exist.\n", testBucketName)
		}
	}

	// 4. 연결 테스트
	fmt.Println("\n4. Connection test...")
	_, err = client.ListBuckets(ctx)
	if err != nil {
		log.Printf("Connection test failed: %v", err)
	} else {
		fmt.Println("✅ MinIO client is working properly.")
	}

	fmt.Println("\n=== MinIO Client Example Completed ===")
}
