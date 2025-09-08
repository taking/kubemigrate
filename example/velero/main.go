package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taking/kubemigrate/pkg/client/velero"
)

func main() {
	fmt.Println("=== Velero Client Usage Example ===")

	// 1. Velero 클라이언트 생성
	fmt.Println("\n1. Creating Velero client...")
	client := velero.NewClient()
	ctx := context.Background()

	// 클라이언트 생성 후 간단한 테스트로 연결 확인
	_, err := client.GetBackups(ctx, "velero")
	if err != nil {
		log.Printf("Velero client connection failed: %v", err)
		fmt.Println("Cannot connect to Velero server. Please check your configuration.")
		fmt.Println("Things to check:")
		fmt.Println("  - Verify that Velero is installed in the cluster")
		fmt.Println("  - Set VELERO_NAMESPACE environment variable (default: velero)")
		fmt.Println("  - Check Velero Pod status with: kubectl get pods -n velero")
		return
	}
	fmt.Println("✅ Velero client created successfully.")

	// 2. 백업 목록 조회
	fmt.Println("\n2. Retrieving backup list...")
	namespace := "velero"
	backups, err := client.GetBackups(ctx, namespace)
	if err != nil {
		log.Printf("Failed to retrieve backup list: %v", err)
	} else {
		fmt.Printf("✅ Found %d backups in '%s' namespace:\n", len(backups), namespace)
		for i, backup := range backups {
			if i < 5 { // 처음 5개만 출력
				fmt.Printf("  - %s (status: %s)\n",
					backup.Name,
					backup.Status.Phase)
			}
		}
		if len(backups) > 5 {
			fmt.Printf("  ... and %d more\n", len(backups)-5)
		}
	}

	// 3. 백업 저장소 목록 조회
	fmt.Println("\n3. Retrieving backup repository list...")
	repositories, err := client.GetBackupRepositories(ctx, namespace)
	if err != nil {
		log.Printf("Failed to retrieve backup repository list: %v", err)
	} else {
		fmt.Printf("✅ Found %d backup repositories in '%s' namespace:\n", len(repositories), namespace)
		for i, repo := range repositories {
			if i < 5 { // 처음 5개만 출력
				fmt.Printf("  - %s (status: %s)\n",
					repo.Name,
					repo.Status.Phase)
			}
		}
		if len(repositories) > 5 {
			fmt.Printf("  ... and %d more\n", len(repositories)-5)
		}
	}

	// 4. 연결 테스트
	fmt.Println("\n4. Connection test...")
	_, err = client.GetBackups(ctx, "velero")
	if err != nil {
		log.Printf("Connection test failed: %v", err)
	} else {
		fmt.Println("✅ Velero client is working properly.")
	}

	fmt.Println("\n=== Velero Client Example Completed ===")
}
