package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/client/helm"
	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/client/velero"
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

	// 모의 클라이언트 생성
	mockClient := &MockClient{}

	// 캐시된 클라이언트 생성
	cached := &CachedClient{
		Client:    mockClient,
		CreatedAt: time.Now(),
		TTL:       ttl,
	}

	// 즉시 만료 확인
	if cached.IsExpired() {
		t.Error("CachedClient should not be expired immediately")
	}

	// 만료 대기
	time.Sleep(ttl + 10*time.Millisecond)

	// 만료 확인
	if !cached.IsExpired() {
		t.Error("CachedClient should be expired after TTL")
	}
}

// TestClientCache_Get - 캐시에서 클라이언트 조회 테스트
// 캐시에서 클라이언트를 올바르게 조회하는지 확인
func TestClientCache_Get(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	// 존재하지 않는 클라이언트 조회
	client, exists := cache.Get("non-existent")
	if exists {
		t.Error("Get() should return false for non-existent client")
	}
	if client != nil {
		t.Error("Get() should return nil client for non-existent key")
	}

	// 클라이언트 추가
	mockClient := &MockClient{}
	cache.Set("test-key", mockClient)

	// 존재하는 클라이언트 조회
	client, exists = cache.Get("test-key")
	if !exists {
		t.Error("Get() should return true for existing client")
	}
	if client != mockClient {
		t.Error("Get() should return the correct client")
	}
}

// TestClientCache_Set - 캐시에 클라이언트 저장 테스트
// 캐시에 클라이언트를 올바르게 저장하는지 확인
func TestClientCache_Set(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	mockClient := &MockClient{}

	// 클라이언트 저장
	cache.Set("test-key", mockClient)

	// 저장 확인
	if len(cache.clients) != 1 {
		t.Errorf("Expected 1 client in cache, got %d", len(cache.clients))
	}

	// 조회 확인
	client, exists := cache.Get("test-key")
	if !exists {
		t.Error("Set() should store the client")
	}
	if client != mockClient {
		t.Error("Set() should store the correct client")
	}
}

// TestClientCache_Get_Expired - 만료된 클라이언트 조회 테스트
// 만료된 클라이언트는 조회되지 않는지 확인
func TestClientCache_Get_Expired(t *testing.T) {
	ttl := 50 * time.Millisecond
	cache := NewClientCache(ttl)

	mockClient := &MockClient{}
	cache.Set("test-key", mockClient)

	// 만료 대기
	time.Sleep(ttl + 10*time.Millisecond)

	// 만료된 클라이언트 조회 불가 확인
	client, exists := cache.Get("test-key")
	if exists {
		t.Error("Get() should return false for expired client")
	}
	if client != nil {
		t.Error("Get() should return nil for expired client")
	}
}

// TestClientCache_Cleanup - 캐시 정리 테스트
// 만료된 클라이언트들이 올바르게 정리되는지 확인
func TestClientCache_Cleanup(t *testing.T) {
	ttl := 50 * time.Millisecond
	cache := NewClientCache(ttl)

	// 클라이언트 추가
	mockClient1 := &MockClient{}
	mockClient2 := &MockClient{}
	cache.Set("key1", mockClient1)
	cache.Set("key2", mockClient2)

	// 저장 확인
	if len(cache.clients) != 2 {
		t.Errorf("Expected 2 clients in cache, got %d", len(cache.clients))
	}

	// 만료 대기
	time.Sleep(ttl + 10*time.Millisecond)

	// 정리 실행
	cache.Cleanup()

	// 제거 확인
	if len(cache.clients) != 0 {
		t.Errorf("Expected 0 clients after cleanup, got %d", len(cache.clients))
	}
}

// TestClientCache_GetOrCreate - 클라이언트 조회 또는 생성 테스트
// 캐시에서 클라이언트를 조회하거나 없으면 새로 생성하는지 확인
func TestClientCache_GetOrCreate(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	kubeConfig := config.KubeConfig{
		KubeConfig: "apiVersion: v1\nkind: Config",
		Namespace:  "default",
	}
	helmConfig := config.KubeConfig{
		KubeConfig: "apiVersion: v1\nkind: Config",
		Namespace:  "default",
	}
	veleroConfig := config.VeleroConfig{
		KubeConfig: kubeConfig,
		MinioConfig: config.MinioConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin123",
			UseSSL:    false,
		},
	}
	minioConfig := config.MinioConfig{
		Endpoint:  "localhost:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin123",
		UseSSL:    false,
	}

	createCount := 0
	createFunc := func() client.Client {
		createCount++
		return &MockClient{}
	}

	// 첫 번째 호출 (생성)
	client1 := cache.GetOrCreate(kubeConfig, helmConfig, veleroConfig, minioConfig, createFunc)
	if client1 == nil {
		t.Error("GetOrCreate() should return a client")
	}
	if createCount != 1 {
		t.Errorf("Expected createFunc to be called 1 time, got %d", createCount)
	}

	// 두 번째 호출 (캐시)
	client2 := cache.GetOrCreate(kubeConfig, helmConfig, veleroConfig, minioConfig, createFunc)
	if client2 == nil {
		t.Error("GetOrCreate() should return a client")
	}
	if createCount != 1 {
		t.Errorf("Expected createFunc to be called 1 time total, got %d", createCount)
	}
	if client1 != client2 {
		t.Error("GetOrCreate() should return the same cached client")
	}
}

// TestClientCache_Stats - 캐시 통계 테스트
// 캐시 통계 정보가 올바르게 계산되는지 확인
func TestClientCache_Stats(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	// 빈 캐시 통계
	stats := cache.Stats()
	if stats["total_clients"] != 0 {
		t.Errorf("Expected 0 total clients, got %v", stats["total_clients"])
	}
	if stats["active_clients"] != 0 {
		t.Errorf("Expected 0 active clients, got %v", stats["active_clients"])
	}
	if stats["expired_clients"] != 0 {
		t.Errorf("Expected 0 expired clients, got %v", stats["expired_clients"])
	}
	if stats["cache_ttl"] != ttl.String() {
		t.Errorf("Expected cache_ttl %v, got %v", ttl.String(), stats["cache_ttl"])
	}

	// 클라이언트 추가
	mockClient1 := &MockClient{}
	mockClient2 := &MockClient{}
	cache.Set("key1", mockClient1)
	cache.Set("key2", mockClient2)

	// 활성 클라이언트 통계
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
}

// TestClientCache_Concurrent - 동시성 테스트
// 여러 고루틴에서 동시에 캐시에 접근해도 안전한지 확인
func TestClientCache_Concurrent(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	// 동시 접근 테스트
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()

			key := fmt.Sprintf("key-%d", i)
			mockClient := &MockClient{}

			// 클라이언트 설정
			cache.Set(key, mockClient)

			// 클라이언트 조회
			client, exists := cache.Get(key)
			if !exists {
				t.Errorf("Client %s should exist", key)
			}
			if client != mockClient {
				t.Errorf("Client %s should match", key)
			}
		}(i)
	}

	// 완료 대기
	for i := 0; i < 10; i++ {
		<-done
	}

	// 저장 확인
	stats := cache.Stats()
	if stats["total_clients"] != 10 {
		t.Errorf("Expected 10 total clients, got %v", stats["total_clients"])
	}
}

// TestClientCache_GenerateCacheKey - 캐시 키 생성 테스트
// 동일한 입력에 대해 동일한 키가 생성되는지 확인
func TestClientCache_GenerateCacheKey(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewClientCache(ttl)

	// 동일한 키 생성 확인
	key1 := cache.generateCacheKey("test-config")
	key2 := cache.generateCacheKey("test-config")

	if key1 != key2 {
		t.Error("Same input should generate same cache key")
	}

	// 다른 키 생성 확인
	key3 := cache.generateCacheKey("different-config")

	if key1 == key3 {
		t.Error("Different input should generate different cache key")
	}

	// 키 비어있지 않음 확인
	if key1 == "" {
		t.Error("Cache key should not be empty")
	}
}

// TestClientCache_ExpiredClientsInStats - 만료된 클라이언트 통계 테스트
// 만료된 클라이언트들이 통계에 올바르게 반영되는지 확인
func TestClientCache_ExpiredClientsInStats(t *testing.T) {
	ttl := 50 * time.Millisecond
	cache := NewClientCache(ttl)

	// 클라이언트 추가
	mockClient := &MockClient{}
	cache.Set("key1", mockClient)

	// 만료 대기
	time.Sleep(ttl + 10*time.Millisecond)

	// 활성 클라이언트 추가
	mockClient2 := &MockClient{}
	cache.Set("key2", mockClient2)

	// 통계 확인
	stats := cache.Stats()
	if stats["total_clients"] != 2 {
		t.Errorf("Expected 2 total clients, got %v", stats["total_clients"])
	}
	if stats["active_clients"] != 1 {
		t.Errorf("Expected 1 active client, got %v", stats["active_clients"])
	}
	if stats["expired_clients"] != 1 {
		t.Errorf("Expected 1 expired client, got %v", stats["expired_clients"])
	}
}

// MockClient - 테스트용 모의 클라이언트
// 테스트에서 사용할 수 있는 클라이언트 인터페이스 구현체
type MockClient struct{}

func (m *MockClient) Kubernetes() kubernetes.Client {
	return kubernetes.NewClient()
}

func (m *MockClient) Helm() helm.Client {
	return helm.NewClient()
}

func (m *MockClient) Velero() velero.Client {
	return velero.NewClient()
}

func (m *MockClient) Minio() minio.Client {
	return minio.NewClient()
}
