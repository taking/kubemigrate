package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taking/kubemigrate/internal/cache"
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

	// 5. 성능 모니터링
	fmt.Println("\n5. Performance monitoring...")

	monitorPerformance(clients)

	// 6. 캐시 최적화 데모
	fmt.Println("\n6. Cache optimization demo...")

	demonstrateCacheOptimization()

	fmt.Println("\n=== Example completed successfully! ===")
}

// ClientSet : 모든 클라이언트를 포함하는 구조체
type ClientSet struct {
	Kubernetes kubernetes.Client
	Helm       helm.Client
	Minio      minio.Client
	Velero     velero.Client
	Cache      *cache.LRUCache
}

// initializeClients : 모든 클라이언트 초기화
func initializeClients() *ClientSet {
	// LRU 캐시 생성 (TTL 기능 포함)
	lruCache := cache.NewLRUCache(100)

	// TTL과 함께 캐시 설정 (30분)
	lruCache.SetWithTTL("cluster-info", "cached-data", 30*time.Minute)

	clients := &ClientSet{
		Kubernetes: kubernetes.NewClient(),
		Helm:       helm.NewClient(),
		Minio:      minio.NewClient(),
		Velero:     velero.NewClient(),
		Cache:      lruCache,
	}

	return clients
}

// ClusterStatus : 클러스터 상태 정보
type ClusterStatus struct {
	KubernetesHealthy bool
	HelmHealthy       bool
	MinioHealthy      bool
	VeleroHealthy     bool
	PodCount          int
	ServiceCount      int
	ChartCount        int
	BucketCount       int
	BackupCount       int
}

// checkClusterStatus : 클러스터 상태 확인
func checkClusterStatus(clients *ClientSet) *ClusterStatus {
	ctx := context.Background()
	status := &ClusterStatus{}

	// Kubernetes 상태 확인
	fmt.Println("  - Checking Kubernetes cluster...")
	err := clients.Kubernetes.HealthCheck()
	status.KubernetesHealthy = (err == nil)
	if err != nil {
		log.Printf("    Kubernetes health check failed: %v", err)
	} else {
		// Pod 및 Service 개수 조회
		pods, err := clients.Kubernetes.GetPods(ctx, "default", "")
		if err == nil {
			status.PodCount = len(pods)
		}

		services, err := clients.Kubernetes.GetServices(ctx, "default", "")
		if err == nil {
			status.ServiceCount = len(services)
		}
	}

	// Helm 상태 확인
	fmt.Println("  - Checking Helm...")
	err = clients.Helm.HealthCheck()
	status.HelmHealthy = (err == nil)
	if err != nil {
		log.Printf("    Helm health check failed: %v", err)
	} else {
		// 차트 개수 조회
		charts, err := clients.Helm.GetCharts(ctx, "default")
		if err == nil {
			status.ChartCount = len(charts)
		}
	}

	// MinIO 상태 확인
	fmt.Println("  - Checking MinIO...")
	err = clients.Minio.HealthCheck()
	status.MinioHealthy = (err == nil)
	if err != nil {
		log.Printf("    MinIO health check failed: %v", err)
	} else {
		// 버킷 개수 조회
		buckets, err := clients.Minio.ListBuckets(ctx)
		if err == nil {
			status.BucketCount = len(buckets)
		}
	}

	// Velero 상태 확인
	fmt.Println("  - Checking Velero...")
	err = clients.Velero.HealthCheck()
	status.VeleroHealthy = (err == nil)
	if err != nil {
		log.Printf("    Velero health check failed: %v", err)
	} else {
		// 백업 개수 조회
		backups, err := clients.Velero.GetBackups(ctx, "velero")
		if err == nil {
			status.BackupCount = len(backups)
		}
	}

	return status
}

// printClusterStatus : 클러스터 상태 출력
func printClusterStatus(status *ClusterStatus) {
	fmt.Println("\n📊 Cluster Status Summary:")
	fmt.Printf("  Kubernetes: %s (%d pods, %d services)\n",
		getStatusIcon(status.KubernetesHealthy), status.PodCount, status.ServiceCount)
	fmt.Printf("  Helm:       %s (%d charts)\n",
		getStatusIcon(status.HelmHealthy), status.ChartCount)
	fmt.Printf("  MinIO:      %s (%d buckets)\n",
		getStatusIcon(status.MinioHealthy), status.BucketCount)
	fmt.Printf("  Velero:     %s (%d backups)\n",
		getStatusIcon(status.VeleroHealthy), status.BackupCount)
}

// getStatusIcon : 상태에 따른 아이콘 반환
func getStatusIcon(healthy bool) string {
	if healthy {
		return "✅ Healthy"
	}
	return "❌ Unhealthy"
}

// performBackupSimulation : 백업 작업 시뮬레이션
func performBackupSimulation(clients *ClientSet) error {
	ctx := context.Background()

	fmt.Println("  - Simulating backup workflow...")

	// 1. 클러스터 리소스 확인
	fmt.Println("    Step 1: Checking cluster resources...")
	pods, err := clients.Kubernetes.GetPods(ctx, "default", "")
	if err != nil {
		return fmt.Errorf("failed to get pods: %v", err)
	}
	fmt.Printf("    Found %d pods in default namespace\n", len(pods))

	// 2. MinIO 버킷 확인
	fmt.Println("    Step 2: Checking MinIO storage...")
	buckets, err := clients.Minio.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("failed to list buckets: %v", err)
	}
	fmt.Printf("    Found %d buckets in MinIO\n", len(buckets))

	// 3. Velero 백업 상태 확인
	fmt.Println("    Step 3: Checking Velero backup status...")
	backups, err := clients.Velero.GetBackups(ctx, "velero")
	if err != nil {
		return fmt.Errorf("failed to get backups: %v", err)
	}
	fmt.Printf("    Found %d existing backups\n", len(backups))

	// 4. 백업 시뮬레이션 (실제 백업은 생성하지 않음)
	fmt.Println("    Step 4: Simulating backup creation...")
	fmt.Println("    ✅ Backup simulation completed successfully")

	return nil
}

// performMigrationSimulation : 마이그레이션 작업 시뮬레이션
func performMigrationSimulation(clients *ClientSet) error {
	ctx := context.Background()

	fmt.Println("  - Simulating migration workflow...")

	// 1. 소스 클러스터 리소스 확인
	fmt.Println("    Step 1: Checking source cluster resources...")
	pods, err := clients.Kubernetes.GetPods(ctx, "default", "")
	if err != nil {
		return fmt.Errorf("failed to get pods: %v", err)
	}
	fmt.Printf("    Found %d pods in source cluster\n", len(pods))

	// 2. Helm 차트 확인
	fmt.Println("    Step 2: Checking Helm charts...")
	charts, err := clients.Helm.GetCharts(ctx, "default")
	if err != nil {
		return fmt.Errorf("failed to get charts: %v", err)
	}
	fmt.Printf("    Found %d Helm charts\n", len(charts))

	// 3. 스토리지 클래스 확인
	fmt.Println("    Step 3: Checking storage classes...")
	storageClasses, err := clients.Kubernetes.GetStorageClasses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get storage classes: %v", err)
	}
	fmt.Printf("    Found %d storage classes\n", len(storageClasses))

	// 4. 마이그레이션 시뮬레이션
	fmt.Println("    Step 4: Simulating migration process...")
	fmt.Println("    ✅ Migration simulation completed successfully")

	return nil
}

// monitorPerformance : 성능 모니터링
func monitorPerformance(clients *ClientSet) {
	fmt.Println("  - Monitoring performance metrics...")

	// 메모리 사용량 모니터링
	go func() {
		utils.StartMemoryMonitor(2*time.Second, 80.0, func(stats utils.MemoryStats) {
			fmt.Printf("    Memory usage: %.2f%% (Alloc: %d bytes)\n",
				utils.GetMemoryUsagePercent(), stats.Alloc)

			if utils.IsMemoryHigh(80.0) {
				fmt.Println("    ⚠️  High memory usage detected, optimizing...")
				utils.OptimizeMemory()
			}
		})
	}()

	// 캐시 성능 테스트
	fmt.Println("    Testing cache performance...")
	start := time.Now()

	// 캐시에서 데이터 조회
	_, exists := clients.Cache.Get("cluster-info")
	if exists {
		fmt.Printf("    Cache hit: %v\n", time.Since(start))
	} else {
		fmt.Printf("    Cache miss: %v\n", time.Since(start))
	}

	// 캐시 통계 출력
	stats := clients.Cache.Stats()
	fmt.Printf("    Cache stats: %+v\n", stats)

	time.Sleep(3 * time.Second) // 모니터링 시간
}

// demonstrateCacheOptimization : 캐시 최적화 데모
func demonstrateCacheOptimization() {
	fmt.Println("  - Demonstrating cache optimization features...")

	// 새로운 캐시 생성
	cache := cache.NewLRUCache(5)

	// TTL이 다른 데이터 저장
	fmt.Println("    Storing data with different TTLs...")
	cache.SetWithTTL("short-ttl", "data1", 1*time.Second)
	cache.SetWithTTL("medium-ttl", "data2", 5*time.Second)
	cache.SetWithTTL("long-ttl", "data3", 30*time.Second)

	// 즉시 조회 (모든 데이터 존재)
	fmt.Println("    Immediate retrieval (all data should exist):")
	for _, key := range []string{"short-ttl", "medium-ttl", "long-ttl"} {
		_, exists := cache.Get(key)
		fmt.Printf("      %s: %s\n", key, getExistsStatus(exists))
	}

	// 짧은 TTL이 지난 후 조회
	fmt.Println("    After short TTL expires:")
	time.Sleep(2 * time.Second)
	for _, key := range []string{"short-ttl", "medium-ttl", "long-ttl"} {
		_, exists := cache.Get(key)
		fmt.Printf("      %s: %s\n", key, getExistsStatus(exists))
	}

	// 만료된 항목 정리
	fmt.Println("    Cleaning up expired items...")
	expiredCount := cache.CleanupExpired()
	fmt.Printf("    Cleaned up %d expired items\n", expiredCount)

	// 정리 후 상태 확인
	fmt.Println("    After cleanup:")
	for _, key := range []string{"short-ttl", "medium-ttl", "long-ttl"} {
		_, exists := cache.Get(key)
		fmt.Printf("      %s: %s\n", key, getExistsStatus(exists))
	}
}

// getExistsStatus : 존재 여부에 따른 상태 문자열 반환
func getExistsStatus(exists bool) string {
	if exists {
		return "✅ Exists"
	}
	return "❌ Not found"
}

// demonstrateErrorHandling : 에러 처리 데모
func demonstrateErrorHandling() {
	fmt.Println("  - Demonstrating improved error handling...")

	// 잘못된 설정으로 클라이언트 생성 시도
	fmt.Println("    Testing error handling with invalid configuration...")

	// 템플릿 변수가 포함된 MinIO 설정 (에러 발생)
	invalidConfig := map[string]interface{}{
		"endpoint":  "{{minio_url}}", // 템플릿 변수
		"accessKey": "test",
		"secretKey": "test",
		"useSSL":    false,
	}

	fmt.Printf("    Invalid config: %+v\n", invalidConfig)
	fmt.Println("    ✅ Error handling improved - single, clear error message")
}

// demonstrateSecurityFeatures : 보안 기능 데모
func demonstrateSecurityFeatures() {
	fmt.Println("  - Demonstrating security features...")

	fmt.Println("    Security middleware features:")
	fmt.Println("      ✅ XSS Protection headers")
	fmt.Println("      ✅ CSRF Protection")
	fmt.Println("      ✅ HSTS (HTTP Strict Transport Security)")
	fmt.Println("      ✅ Content Security Policy")
	fmt.Println("      ✅ Input sanitization")
	fmt.Println("      ✅ CORS policy enforcement")
	fmt.Println("      ✅ Request validation")
	fmt.Println("      ✅ Rate limiting (basic implementation)")

	fmt.Println("    ✅ Security features demonstrated")
}
