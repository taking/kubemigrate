package cache

import (
	"testing"
	"time"

	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/pkg/client"
)

// TestNewClientCache - 클라이언트 캐시 생성자 테스트
// 새로 생성된 캐시가 올바르게 초기화되었는지 확인
func TestNewClientCache(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	if cache == nil {
		t.Fatal("NewClientCache() returned nil")
	}

	if cache.ttl != ttl {
		t.Errorf("NewClientCache() ttl = %v, want %v", cache.ttl, ttl)
	}

	if cache.clients == nil {
		t.Error("NewClientCache() clients map is nil")
	}

	if len(cache.clients) != 0 {
		t.Errorf("NewClientCache() clients map should be empty, got %d", len(cache.clients))
	}
}

// TestCachedClient_IsExpired - 캐시된 클라이언트 만료 테스트
// TTL에 따른 클라이언트 만료 처리 확인
func TestCachedClient_IsExpired(t *testing.T) {
	ttl := 100 * time.Millisecond

	// 실제 클라이언트 생성
	mockClient := client.NewClient()

	// 캐시된 클라이언트 생성
	cached := &CachedClient{
		Client:    mockClient,
		ApiType:   "test",
		CreatedAt: time.Now(),
		TTL:       ttl,
	}

	// 즉시 만료 확인
	if cached.IsExpired() {
		t.Error("CachedClient should not be expired immediately")
	}

	// TTL 후 만료 확인
	time.Sleep(ttl + 10*time.Millisecond)
	if !cached.IsExpired() {
		t.Error("CachedClient should be expired after TTL")
	}
}

// TestClientCache_Get - 캐시에서 클라이언트 조회 테스트
// 캐시에서 클라이언트를 올바르게 조회하는지 확인
func TestClientCache_Get(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	// 테스트용 설정 생성
	kubeConfig := config.KubeConfig{Config: "test-kubeconfig"}
	helmConfig := config.KubeConfig{Config: "test-helmconfig"}
	veleroConfig := config.VeleroConfig{}
	minioConfig := config.MinioConfig{Endpoint: "test-endpoint"}

	// 존재하지 않는 클라이언트 조회
	clientResult, exists := cache.Get("minio", kubeConfig, helmConfig, veleroConfig, minioConfig)
	if exists {
		t.Error("Get() should return false for non-existent key")
	}
	if clientResult != nil {
		t.Error("Get() should return nil client for non-existent key")
	}

	// 클라이언트 추가
	mockClient := client.NewClient()
	cache.Set("minio", kubeConfig, helmConfig, veleroConfig, minioConfig, mockClient)

	// 존재하는 클라이언트 조회
	clientResult, exists = cache.Get("minio", kubeConfig, helmConfig, veleroConfig, minioConfig)
	if !exists {
		t.Error("Get() should return true for existing client")
	}
	if clientResult != mockClient {
		t.Error("Get() should return the correct client")
	}
}

// TestClientCache_Set - 캐시에 클라이언트 저장 테스트
// 캐시에 클라이언트를 올바르게 저장하는지 확인
func TestClientCache_Set(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	mockClient := client.NewClient()
	kubeConfig := config.KubeConfig{Config: "test-kubeconfig"}
	helmConfig := config.KubeConfig{Config: "test-helmconfig"}
	veleroConfig := config.VeleroConfig{}
	minioConfig := config.MinioConfig{Endpoint: "test-endpoint"}

	// 클라이언트 저장
	cache.Set("minio", kubeConfig, helmConfig, veleroConfig, minioConfig, mockClient)

	// 저장 확인
	if len(cache.clients) != 1 {
		t.Errorf("Expected 1 client in cache, got %d", len(cache.clients))
	}

	// 조회 확인
	clientResult, exists := cache.Get("minio", kubeConfig, helmConfig, veleroConfig, minioConfig)
	if !exists {
		t.Error("Set() should store the client")
	}
	if clientResult != mockClient {
		t.Error("Set() should store the correct client")
	}
}

// TestClientCache_Get_Expired - 만료된 클라이언트 조회 테스트
// 만료된 클라이언트는 조회되지 않는지 확인
func TestClientCache_Get_Expired(t *testing.T) {
	ttl := 50 * time.Millisecond
	cache := NewClientCache(ttl)

	mockClient := client.NewClient()
	kubeConfig := config.KubeConfig{Config: "test-kubeconfig"}
	helmConfig := config.KubeConfig{Config: "test-helmconfig"}
	veleroConfig := config.VeleroConfig{}
	minioConfig := config.MinioConfig{Endpoint: "test-endpoint"}

	// 클라이언트 저장
	cache.Set("minio", kubeConfig, helmConfig, veleroConfig, minioConfig, mockClient)

	// TTL 후 조회
	time.Sleep(ttl + 10*time.Millisecond)
	clientResult, exists := cache.Get("minio", kubeConfig, helmConfig, veleroConfig, minioConfig)

	if exists {
		t.Error("Get() should return false for expired client")
	}
	if clientResult != nil {
		t.Error("Get() should return nil for expired client")
	}
}

// TestClientCache_GetOrCreate - 캐시에서 조회하거나 새로 생성 테스트
// 캐시에 없으면 새로 생성하는지 확인
func TestClientCache_GetOrCreate(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	kubeConfig := config.KubeConfig{Config: "test-kubeconfig"}
	helmConfig := config.KubeConfig{Config: "test-helmconfig"}
	veleroConfig := config.VeleroConfig{}
	minioConfig := config.MinioConfig{Endpoint: "test-endpoint"}

	// 첫 번째 호출 - 새로 생성
	client1 := cache.GetOrCreate("minio", kubeConfig, helmConfig, veleroConfig, minioConfig, func() client.Client {
		return client.NewClient()
	})

	if client1 == nil {
		t.Error("GetOrCreate() should return a client")
	}

	// 두 번째 호출 - 캐시에서 조회
	client2 := cache.GetOrCreate("minio", kubeConfig, helmConfig, veleroConfig, minioConfig, func() client.Client {
		t.Error("CreateFunc should not be called for cached client")
		return client.NewClient()
	})

	if client1 != client2 {
		t.Error("GetOrCreate() should return the same client from cache")
	}
}

// TestClientCache_Cleanup - 만료된 캐시 정리 테스트
// 만료된 클라이언트가 정리되는지 확인
func TestClientCache_Cleanup(t *testing.T) {
	ttl := 50 * time.Millisecond
	cache := NewClientCache(ttl)

	kubeConfig1 := config.KubeConfig{Config: "test-kubeconfig-1"}
	kubeConfig2 := config.KubeConfig{Config: "test-kubeconfig-2"}
	helmConfig := config.KubeConfig{Config: "test-helmconfig"}
	veleroConfig := config.VeleroConfig{}
	minioConfig := config.MinioConfig{Endpoint: "test-endpoint"}

	// 두 개의 클라이언트 저장
	cache.Set("minio", kubeConfig1, helmConfig, veleroConfig, minioConfig, client.NewClient())
	cache.Set("kubernetes", kubeConfig2, helmConfig, veleroConfig, minioConfig, client.NewClient())

	if len(cache.clients) != 2 {
		t.Errorf("Expected 2 clients in cache, got %d", len(cache.clients))
	}

	// TTL 후 정리
	time.Sleep(ttl + 10*time.Millisecond)
	cache.Cleanup()

	if len(cache.clients) != 0 {
		t.Errorf("Expected 0 clients after cleanup, got %d", len(cache.clients))
	}
}

// TestClientCache_Invalidate - 특정 설정의 캐시 무효화 테스트
// 특정 설정의 캐시가 무효화되는지 확인
func TestClientCache_Invalidate(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	kubeConfig1 := config.KubeConfig{Config: "test-kubeconfig-1"}
	kubeConfig2 := config.KubeConfig{Config: "test-kubeconfig-2"}
	helmConfig := config.KubeConfig{Config: "test-helmconfig"}
	veleroConfig := config.VeleroConfig{}
	minioConfig := config.MinioConfig{Endpoint: "test-endpoint"}

	// 두 개의 클라이언트 저장
	cache.Set("minio", kubeConfig1, helmConfig, veleroConfig, minioConfig, client.NewClient())
	cache.Set("kubernetes", kubeConfig2, helmConfig, veleroConfig, minioConfig, client.NewClient())

	if len(cache.clients) != 2 {
		t.Errorf("Expected 2 clients in cache, got %d", len(cache.clients))
	}

	// 첫 번째 클라이언트만 무효화
	cache.Invalidate("minio", kubeConfig1, helmConfig, veleroConfig, minioConfig)

	if len(cache.clients) != 1 {
		t.Errorf("Expected 1 client after invalidate, got %d", len(cache.clients))
	}

	// 두 번째 클라이언트는 여전히 존재
	_, exists := cache.Get("kubernetes", kubeConfig2, helmConfig, veleroConfig, minioConfig)
	if !exists {
		t.Error("Second client should still exist after invalidating first")
	}
}

// TestClientCache_InvalidateAll - 모든 캐시 무효화 테스트
// 모든 캐시가 무효화되는지 확인
func TestClientCache_InvalidateAll(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	kubeConfig1 := config.KubeConfig{Config: "test-kubeconfig-1"}
	kubeConfig2 := config.KubeConfig{Config: "test-kubeconfig-2"}
	helmConfig := config.KubeConfig{Config: "test-helmconfig"}
	veleroConfig := config.VeleroConfig{}
	minioConfig := config.MinioConfig{Endpoint: "test-endpoint"}

	// 두 개의 클라이언트 저장
	cache.Set("minio", kubeConfig1, helmConfig, veleroConfig, minioConfig, client.NewClient())
	cache.Set("kubernetes", kubeConfig2, helmConfig, veleroConfig, minioConfig, client.NewClient())

	if len(cache.clients) != 2 {
		t.Errorf("Expected 2 clients in cache, got %d", len(cache.clients))
	}

	// 모든 캐시 무효화
	cache.InvalidateAll()

	if len(cache.clients) != 0 {
		t.Errorf("Expected 0 clients after invalidate all, got %d", len(cache.clients))
	}
}

// TestClientCache_Stats - 캐시 통계 정보 테스트
// 캐시 통계가 올바르게 반환되는지 확인
func TestClientCache_Stats(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	kubeConfig1 := config.KubeConfig{Config: "test-kubeconfig-1"}
	kubeConfig2 := config.KubeConfig{Config: "test-kubeconfig-2"}
	helmConfig := config.KubeConfig{Config: "test-helmconfig"}
	veleroConfig := config.VeleroConfig{}
	minioConfig := config.MinioConfig{Endpoint: "test-endpoint"}

	// 빈 캐시 통계
	stats := cache.Stats()
	if stats["total_clients"] != 0 {
		t.Errorf("Expected 0 total clients, got %v", stats["total_clients"])
	}

	// 클라이언트 추가
	cache.Set("minio", kubeConfig1, helmConfig, veleroConfig, minioConfig, client.NewClient())
	cache.Set("kubernetes", kubeConfig2, helmConfig, veleroConfig, minioConfig, client.NewClient())

	// 통계 확인
	stats = cache.Stats()
	if stats["total_clients"] != 2 {
		t.Errorf("Expected 2 total clients, got %v", stats["total_clients"])
	}
	if stats["active_clients"] != 2 {
		t.Errorf("Expected 2 active clients, got %v", stats["active_clients"])
	}
	if stats["expired_clients"] != 0 {
		t.Errorf("Expected 0 expired clients, got %v", stats["expired_clients"])
	}

	// API 타입별 통계 확인
	apiTypeStats, ok := stats["api_type_stats"].(map[string]int)
	if !ok {
		t.Error("api_type_stats should be present")
	}
	if apiTypeStats["minio"] != 1 {
		t.Errorf("Expected 1 minio client, got %d", apiTypeStats["minio"])
	}
	if apiTypeStats["kubernetes"] != 1 {
		t.Errorf("Expected 1 kubernetes client, got %d", apiTypeStats["kubernetes"])
	}
}

// TestClientCache_ConcurrentAccess - 동시 접근 테스트
// 여러 고루틴에서 동시에 캐시에 접근해도 안전한지 확인
func TestClientCache_ConcurrentAccess(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	kubeConfig := config.KubeConfig{Config: "test-kubeconfig"}
	helmConfig := config.KubeConfig{Config: "test-helmconfig"}
	veleroConfig := config.VeleroConfig{}
	minioConfig := config.MinioConfig{Endpoint: "test-endpoint"}

	// 클라이언트를 미리 생성하여 동시성 문제 방지
	mockClient1 := client.NewClient()
	mockClient2 := client.NewClient()

	// 동시에 여러 클라이언트 추가
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			apiType := "minio"
			clientToUse := mockClient1
			if id%2 == 0 {
				apiType = "kubernetes"
				clientToUse = mockClient2
			}
			cache.Set(apiType, kubeConfig, helmConfig, veleroConfig, minioConfig, clientToUse)
			done <- true
		}(i)
	}

	// 모든 고루틴 완료 대기
	for i := 0; i < 10; i++ {
		<-done
	}

	// 캐시 상태 확인
	stats := cache.Stats()
	if stats["total_clients"] != 2 {
		t.Errorf("Expected 2 unique clients, got %v", stats["total_clients"])
	}
}
