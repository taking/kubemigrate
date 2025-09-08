package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taking/kubemigrate/pkg/client/helm"
	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/client/velero"
	"github.com/taking/kubemigrate/pkg/utils"
)

func main() {
	fmt.Println("=== KubeMigrate Integrated Usage Example ===")
	fmt.Println("This example demonstrates real-world scenarios using all clients together.")

	// 1. 모든 클라이언트 초기화
	fmt.Println("\n1. Initializing clients...")

	clients := initializeClients()
	if clients == nil {
		fmt.Println("❌ Client initialization failed")
		return
	}

	fmt.Println("✅ All clients initialized successfully.")

	// 2. 클러스터 상태 확인
	fmt.Println("\n2. Checking cluster status...")

	clusterStatus := checkClusterStatus(clients)
	printClusterStatus(clusterStatus)

	// 3. 백업 작업 시뮬레이션
	fmt.Println("\n3. Backup operation simulation...")

	err := performBackupSimulation(clients)
	if err != nil {
		log.Printf("Backup simulation failed: %v", err)
	}

	// 4. 마이그레이션 작업 시뮬레이션
	fmt.Println("\n4. Migration operation simulation...")

	err = performMigrationSimulation(clients)
	if err != nil {
		log.Printf("Migration simulation failed: %v", err)
	}

	// 5. 모니터링 및 알림
	fmt.Println("\n5. Monitoring and notifications...")

	err = performMonitoring(clients)
	if err != nil {
		log.Printf("Monitoring failed: %v", err)
	}

	fmt.Println("\n=== Integrated Example Completed ===")
}

// 클라이언트 구조체
type Clients struct {
	Kubernetes kubernetes.Client
	Helm       helm.Client
	MinIO      minio.Client
	Velero     velero.Client
}

// 클러스터 상태 구조체
type ClusterStatus struct {
	KubernetesHealthy bool
	HelmHealthy       bool
	MinIOHealthy      bool
	VeleroHealthy     bool
	PodCount          int
	ConfigMapCount    int
	ChartCount        int
	BackupCount       int
	BucketCount       int
}

// 클라이언트 초기화
func initializeClients() *Clients {
	clients := &Clients{
		Kubernetes: kubernetes.NewClient(),
		Helm:       helm.NewClient(),
		MinIO:      minio.NewClient(),
		Velero:     velero.NewClient(),
	}

	ctx := context.Background()

	// 각 클라이언트의 연결 테스트
	fmt.Println("  Testing Kubernetes client connection...")
	if _, err := clients.Kubernetes.GetNamespaces(ctx); err != nil {
		log.Printf("    ❌ Kubernetes client: %v", err)
	} else {
		fmt.Println("    ✅ Kubernetes client healthy")
	}

	fmt.Println("  Testing Helm client health...")
	if err := clients.Helm.HealthCheck(ctx); err != nil {
		log.Printf("    ❌ Helm client: %v", err)
	} else {
		fmt.Println("    ✅ Helm client healthy")
	}

	fmt.Println("  Testing MinIO client connection...")
	if _, err := clients.MinIO.ListBuckets(ctx); err != nil {
		log.Printf("    ❌ MinIO client: %v", err)
	} else {
		fmt.Println("    ✅ MinIO client healthy")
	}

	fmt.Println("  Testing Velero client connection...")
	if _, err := clients.Velero.GetBackups(ctx, "velero"); err != nil {
		log.Printf("    ❌ Velero client: %v", err)
	} else {
		fmt.Println("    ✅ Velero client healthy")
	}

	return clients
}

// 클러스터 상태 확인
func checkClusterStatus(clients *Clients) *ClusterStatus {
	status := &ClusterStatus{}
	ctx := context.Background()

	// Kubernetes 상태 확인
	if _, err := clients.Kubernetes.GetNamespaces(ctx); err == nil {
		status.KubernetesHealthy = true

		// Pod 수 조회
		pods, err := clients.Kubernetes.GetPods(ctx, "default")
		if err == nil {
			status.PodCount = len(pods.Items)
		}

		// ConfigMap 수 조회
		configMaps, err := clients.Kubernetes.GetConfigMaps(ctx, "default")
		if err == nil {
			status.ConfigMapCount = len(configMaps.Items)
		}
	}

	// Helm 상태 확인
	if err := clients.Helm.HealthCheck(ctx); err == nil {
		status.HelmHealthy = true

		// 차트 수 조회
		charts, err := clients.Helm.GetCharts(ctx, "default")
		if err == nil {
			status.ChartCount = len(charts)
		}
	}

	// MinIO 상태 확인
	if _, err := clients.MinIO.ListBuckets(ctx); err == nil {
		status.MinIOHealthy = true

		// 버킷 수 조회 (interface{} 타입이므로 타입 어설션 필요)
		buckets, err := clients.MinIO.ListBuckets(ctx)
		if err == nil {
			if bucketList, ok := buckets.([]interface{}); ok {
				status.BucketCount = len(bucketList)
			}
		}
	}

	// Velero 상태 확인
	if _, err := clients.Velero.GetBackups(ctx, "velero"); err == nil {
		status.VeleroHealthy = true

		// 백업 수 조회
		backups, err := clients.Velero.GetBackups(ctx, "velero")
		if err == nil {
			status.BackupCount = len(backups)
		}
	}

	return status
}

// 클러스터 상태 출력
func printClusterStatus(status *ClusterStatus) {
	fmt.Printf("Cluster Status:\n")
	fmt.Printf("  Kubernetes: %v (Pods: %d, ConfigMaps: %d)\n",
		status.KubernetesHealthy, status.PodCount, status.ConfigMapCount)
	fmt.Printf("  Helm: %v (Charts: %d)\n",
		status.HelmHealthy, status.ChartCount)
	fmt.Printf("  MinIO: %v (Buckets: %d)\n",
		status.MinIOHealthy, status.BucketCount)
	fmt.Printf("  Velero: %v (Backups: %d)\n",
		status.VeleroHealthy, status.BackupCount)
}

// 백업 작업 시뮬레이션
func performBackupSimulation(clients *Clients) error {
	fmt.Println("Starting backup operation simulation...")

	// 클러스터 상태 확인
	status := checkClusterStatus(clients)

	// 1. Velero 백업 생성
	if status.VeleroHealthy {
		fmt.Println("  1. Creating Velero backup...")

		backupName := "migration-backup-" + time.Now().Format("20060102-150405")
		fmt.Printf("    Backup name: %s\n", backupName)

		fmt.Println("    ✅ Velero backup creation simulation completed")
	}

	// minio에 백업 메타데이터 저장
	if status.MinIOHealthy {
		fmt.Println("  2. Storing backup metadata to MinIO...")

		bucketName := "backup-metadata"
		objectName := "backup-" + time.Now().Format("20060102-150405") + ".json"

		// 백업 메타데이터 생성
		metadata := map[string]interface{}{
			"backup_name":  "migration-backup-" + time.Now().Format("20060102-150405"),
			"created_at":   time.Now().Format(time.RFC3339),
			"cluster_info": "production-cluster",
			"backup_type":  "full",
			"namespaces":   []string{"default", "kube-system"},
			"resources":    []string{"pods", "services", "configmaps", "secrets"},
		}

		fmt.Printf("    Bucket: %s\n", bucketName)
		fmt.Printf("    Object: %s\n", objectName)
		fmt.Printf("    Metadata: %+v\n", metadata)

		fmt.Println("    ✅ Backup metadata storage simulation completed")
	}

	// 3. 백업 상태 모니터링
	fmt.Println("  3. Monitoring backup status...")

	// 타임아웃을 사용한 모니터링
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := utils.RunWithTimeout(ctx, func() error {
		// 백업 상태 확인 시뮬레이션
		time.Sleep(2 * time.Second)
		fmt.Println("    Backup status: Completed")
		return nil
	})

	if err != nil {
		if err == context.DeadlineExceeded {
			fmt.Println("    ⏰ Backup monitoring timeout")
		} else {
			return fmt.Errorf("backup monitoring failed: %w", err)
		}
	} else {
		fmt.Println("    ✅ Backup monitoring completed")
	}

	fmt.Println("✅ Backup operation simulation completed")
	return nil
}

// 마이그레이션 작업 시뮬레이션
func performMigrationSimulation(clients *Clients) error {
	fmt.Println("Starting migration operation simulation...")

	// 클러스터 상태 확인
	status := checkClusterStatus(clients)

	// 1. 소스 클러스터에서 리소스 정보 수집
	if status.KubernetesHealthy {
		fmt.Println("  1. Collecting resource information from source cluster...")

		ctx := context.Background()

		// Pod 정보 수집
		pods, err := clients.Kubernetes.GetPods(ctx, "default")
		if err == nil {
			fmt.Printf("    Collected Pods: %d\n", len(pods.Items))
		}

		// ConfigMap 정보 수집
		configMaps, err := clients.Kubernetes.GetConfigMaps(ctx, "default")
		if err == nil {
			fmt.Printf("    Collected ConfigMaps: %d\n", len(configMaps.Items))
		}

		fmt.Println("    ✅ Resource information collection completed")
	}

	// 2. Helm 차트 정보 수집
	if status.HelmHealthy {
		fmt.Println("  2. Collecting Helm chart information...")

		ctx := context.Background()
		charts, err := clients.Helm.GetCharts(ctx, "default")
		if err != nil {
			log.Printf("    Chart information collection failed: %v", err)
		} else {
			fmt.Printf("    Collected charts: %d\n", len(charts))

			// 각 차트의 상세 정보 수집
			for i, chart := range charts {
				if i < 3 { // 처음 3개만 출력
					fmt.Printf("      Chart %d: %s (version: %s)\n",
						i+1, chart.Name, chart.Chart.Metadata.Version)
				}
			}
		}

		fmt.Println("    ✅ Helm chart information collection completed")
	}

	// 3. 마이그레이션 계획 생성
	fmt.Println("  3. Creating migration plan...")

	migrationPlan := map[string]interface{}{
		"source_cluster":     "production-cluster",
		"target_cluster":     "staging-cluster",
		"migration_date":     time.Now().Format(time.RFC3339),
		"estimated_duration": "2 hours",
		"resources": map[string]interface{}{
			"pods":            status.PodCount,
			"configmaps":      status.ConfigMapCount,
			"charts":          status.ChartCount,
			"storage_classes": 3, // 시뮬레이션
		},
		"backup_info": map[string]interface{}{
			"backup_name":     "migration-backup-" + time.Now().Format("20060102-150405"),
			"backup_size":     "1.2 GB",
			"backup_location": "s3://backup-bucket/",
		},
	}

	fmt.Printf("    Migration plan: %+v\n", migrationPlan)
	fmt.Println("    ✅ Migration plan creation completed")

	fmt.Println("✅ Migration operation simulation completed")
	return nil
}

// 모니터링 및 알림
func performMonitoring(clients *Clients) error {
	fmt.Println("Starting monitoring and notifications...")

	// 1. 클러스터 상태 모니터링
	fmt.Println("  1. Monitoring cluster status...")

	ctx := context.Background()

	// 각 클라이언트의 상태 확인
	statuses := map[string]bool{
		"Kubernetes": func() bool { _, err := clients.Kubernetes.GetNamespaces(ctx); return err == nil }(),
		"Helm":       clients.Helm.HealthCheck(ctx) == nil,
		"MinIO":      func() bool { _, err := clients.MinIO.ListBuckets(ctx); return err == nil }(),
		"Velero":     func() bool { _, err := clients.Velero.GetBackups(ctx, "velero"); return err == nil }(),
	}

	fmt.Println("    Cluster status:")
	for service, healthy := range statuses {
		if healthy {
			fmt.Printf("      ✅ %s: healthy\n", service)
		} else {
			fmt.Printf("      ❌ %s: unhealthy\n", service)
		}
	}

	// 2. 알림 시뮬레이션
	fmt.Println("  2. Notification simulation...")

	alerts := []struct {
		level   string
		message string
		service string
	}{
		{"INFO", "Migration operation completed successfully", "Migration"},
		{"WARNING", "Backup storage usage exceeded 80%", "Velero"},
		{"INFO", "New Helm chart has been installed", "Helm"},
		{"ERROR", "MinIO connection failed", "MinIO"},
	}

	for _, alert := range alerts {
		fmt.Printf("    [%s] %s: %s\n", alert.level, alert.service, alert.message)
	}

	fmt.Println("✅ Monitoring and notifications completed")
	return nil
}
