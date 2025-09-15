package cache

import (
	"fmt"
	"testing"

	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/client/helm"
	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/client/velero"
	"github.com/taking/kubemigrate/pkg/config"
)

// TestNewLRUCache LRU 캐시 생성자 테스트
func TestNewLRUCache(t *testing.T) {
	capacity := 10
	cache := NewLRUCache(capacity)

	if cache == nil {
		t.Fatal("NewLRUCache() returned nil")
	}

	if cache.capacity != capacity {
		t.Errorf("Expected capacity %d, got %d", capacity, cache.capacity)
	}

	if cache.items == nil {
		t.Error("Items map is nil")
	}

	if cache.list == nil {
		t.Error("List is nil")
	}
}

// TestLRUCache_SetAndGet 기본 설정 및 조회 테스트
func TestLRUCache_SetAndGet(t *testing.T) {
	cache := NewLRUCache(5)
	mockClient := &MockClient{}

	// 설정
	cache.Set("key1", mockClient)

	// 조회
	client, exists := cache.Get("key1")
	if !exists {
		t.Error("Expected client to exist")
	}
	if client != mockClient {
		t.Error("Expected correct client")
	}
}

// TestLRUCache_CapacityOverflow 용량 초과 시 오래된 항목 제거 테스트
func TestLRUCache_CapacityOverflow(t *testing.T) {
	cache := NewLRUCache(2)

	// 3개 항목 추가 (용량 초과)
	cache.Set("key1", &MockClient{})
	cache.Set("key2", &MockClient{})
	cache.Set("key3", &MockClient{})

	// 첫 번째 항목이 제거되어야 함
	_, exists := cache.Get("key1")
	if exists {
		t.Error("Expected key1 to be evicted")
	}

	// 나머지 항목들은 존재해야 함
	_, exists = cache.Get("key2")
	if !exists {
		t.Error("Expected key2 to exist")
	}

	_, exists = cache.Get("key3")
	if !exists {
		t.Error("Expected key3 to exist")
	}
}

// TestLRUCache_Expiration LRU 캐시에서는 만료 개념이 없으므로 제거됨
// LRU 캐시는 용량 제한으로만 관리됩니다.

// TestLRUCache_AccessOrder 접근 순서에 따른 LRU 동작 테스트
func TestLRUCache_AccessOrder(t *testing.T) {
	cache := NewLRUCache(3)

	// 3개 항목 추가
	cache.Set("key1", &MockClient{})
	cache.Set("key2", &MockClient{})
	cache.Set("key3", &MockClient{})

	// key1에 접근 (가장 최근 사용으로 변경)
	cache.Get("key1")

	// key4 추가 (key2가 제거되어야 함)
	cache.Set("key4", &MockClient{})

	// key1과 key3, key4는 존재해야 함
	_, exists := cache.Get("key1")
	if !exists {
		t.Error("Expected key1 to exist")
	}

	_, exists = cache.Get("key3")
	if !exists {
		t.Error("Expected key3 to exist")
	}

	_, exists = cache.Get("key4")
	if !exists {
		t.Error("Expected key4 to exist")
	}

	// key2는 제거되어야 함
	_, exists = cache.Get("key2")
	if exists {
		t.Error("Expected key2 to be evicted")
	}
}

// TestLRUCache_Cleanup 정리 테스트 (LRU 캐시용)
func TestLRUCache_Cleanup(t *testing.T) {
	cache := NewLRUCache(5)

	// 항목 추가
	cache.Set("key1", &MockClient{})
	cache.Set("key2", &MockClient{})

	// 정리 실행 (LRU에서는 전체 캐시를 비움)
	cache.Cleanup()

	// 항목들이 모두 제거되어야 함
	stats := cache.Stats()
	if stats["total_items"] != 0 {
		t.Errorf("Expected 0 total items after cleanup, got %v", stats["total_items"])
	}
}

// TestLRUCache_CleanByKey 특정 키 정리 테스트
func TestLRUCache_CleanByKey(t *testing.T) {
	cache := NewLRUCache(5)

	// 항목 추가
	cache.Set("key1", &MockClient{})
	cache.Set("key2", &MockClient{})

	// 특정 키 정리
	removed := cache.CleanByKey("key1")
	if !removed {
		t.Error("Expected key1 to be removed")
	}

	// key1은 제거되고 key2는 남아있어야 함
	stats := cache.Stats()
	if stats["total_items"] != 1 {
		t.Errorf("Expected 1 total item, got %v", stats["total_items"])
	}

	// 존재하지 않는 키 정리
	removed = cache.CleanByKey("nonexistent")
	if removed {
		t.Error("Expected nonexistent key to not be removed")
	}
}

// TestLRUCache_CleanByPattern 패턴 기반 정리 테스트
func TestLRUCache_CleanByPattern(t *testing.T) {
	cache := NewLRUCache(10)

	// 다양한 패턴의 키 추가
	cache.Set("kubernetes-key1", &MockClient{})
	cache.Set("kubernetes-key2", &MockClient{})
	cache.Set("minio-key1", &MockClient{})
	cache.Set("velero-key1", &MockClient{})

	// kubernetes 패턴으로 정리
	removedCount := cache.CleanByPattern("kubernetes")
	if removedCount != 2 {
		t.Errorf("Expected 2 items to be removed, got %d", removedCount)
	}

	// kubernetes 키들이 제거되고 minio, velero 키는 남아있어야 함
	stats := cache.Stats()
	if stats["total_items"] != 2 {
		t.Errorf("Expected 2 total items, got %v", stats["total_items"])
	}
}

// TestLRUCache_GetOrCreateWithApiType API 타입을 명시적으로 지정한 조회 또는 생성 테스트
func TestLRUCache_GetOrCreateWithApiType(t *testing.T) {
	cache := NewLRUCache(5)

	kubeConfig := config.KubeConfig{
		KubeConfig: "apiVersion: v1\nkind: Config",
	}
	helmConfig := config.KubeConfig{
		KubeConfig: "apiVersion: v1\nkind: Config",
	}
	veleroConfig := config.VeleroConfig{
		KubeConfig: config.KubeConfig{
			KubeConfig: "apiVersion: v1\nkind: Config",
		},
	}
	minioConfig := config.MinioConfig{
		Endpoint:  "minio.example.com",
		AccessKey: "access",
		SecretKey: "secret",
	}

	// Kubernetes API 타입으로 생성
	client1 := cache.GetOrCreateWithApiType(kubeConfig, helmConfig, veleroConfig, minioConfig, "kubernetes", func() client.Client {
		return &MockClient{}
	})

	// Helm API 타입으로 생성 (같은 설정이지만 다른 타입)
	client2 := cache.GetOrCreateWithApiType(kubeConfig, helmConfig, veleroConfig, minioConfig, "helm", func() client.Client {
		return &MockClient{}
	})

	// Velero API 타입으로 생성
	client3 := cache.GetOrCreateWithApiType(kubeConfig, helmConfig, veleroConfig, minioConfig, "velero", func() client.Client {
		return &MockClient{}
	})

	// MinIO API 타입으로 생성
	client4 := cache.GetOrCreateWithApiType(kubeConfig, helmConfig, veleroConfig, minioConfig, "minio", func() client.Client {
		return &MockClient{}
	})

	if client1 == nil || client2 == nil || client3 == nil || client4 == nil {
		t.Error("Expected all clients to be created")
	}

	// 상세 통계 확인
	detailedStats := cache.GetDetailedStats()
	if len(detailedStats.ActiveClients) != 4 {
		t.Errorf("Expected 4 active clients, got %d", len(detailedStats.ActiveClients))
	}

	// API 타입 확인
	apiTypes := make(map[string]bool)
	for _, client := range detailedStats.ActiveClients {
		apiTypes[client.ApiType] = true
	}

	expectedTypes := []string{"kubernetes", "helm", "velero", "minio"}
	for _, expectedType := range expectedTypes {
		if !apiTypes[expectedType] {
			t.Errorf("Expected API type %s to be present", expectedType)
		}
	}
}

// TestLRUCache_GetOrCreate 조회 또는 생성 테스트
func TestLRUCache_GetOrCreate(t *testing.T) {
	cache := NewLRUCache(5)

	kubeConfig := config.KubeConfig{
		KubeConfig: "apiVersion: v1\nkind: Config",
		Namespace:  "default",
	}

	createCount := 0
	createFunc := func() client.Client {
		createCount++
		return &MockClient{}
	}

	// 첫 번째 호출 (생성)
	client1 := cache.GetOrCreate(kubeConfig, kubeConfig, config.VeleroConfig{}, config.MinioConfig{}, createFunc)
	if client1 == nil {
		t.Error("Expected client to be created")
	}
	if createCount != 1 {
		t.Errorf("Expected createFunc to be called 1 time, got %d", createCount)
	}

	// 두 번째 호출 (캐시에서 조회)
	client2 := cache.GetOrCreate(kubeConfig, kubeConfig, config.VeleroConfig{}, config.MinioConfig{}, createFunc)
	if client2 == nil {
		t.Error("Expected client to be retrieved")
	}
	if createCount != 1 {
		t.Errorf("Expected createFunc to be called 1 time total, got %d", createCount)
	}
	if client1 != client2 {
		t.Error("Expected same client instance")
	}
}

// TestLRUCache_Stats 통계 테스트
func TestLRUCache_Stats(t *testing.T) {
	cache := NewLRUCache(5)

	// 빈 캐시 통계
	stats := cache.Stats()
	if stats["total_items"] != 0 {
		t.Errorf("Expected 0 total items, got %v", stats["total_items"])
	}
	if stats["capacity"] != 5 {
		t.Errorf("Expected capacity 5, got %v", stats["capacity"])
	}

	// 항목 추가
	cache.Set("key1", &MockClient{})
	cache.Set("key2", &MockClient{})

	// 통계 확인
	stats = cache.Stats()
	if stats["total_items"] != 2 {
		t.Errorf("Expected 2 total items, got %v", stats["total_items"])
	}
	if stats["active_items"] != 2 {
		t.Errorf("Expected 2 active items, got %v", stats["active_items"])
	}
}

// TestLRUCache_Concurrent 동시성 테스트
func TestLRUCache_Concurrent(t *testing.T) {
	cache := NewLRUCache(10)

	// 동시 접근 테스트
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()

			key := fmt.Sprintf("key-%d", i)
			mockClient := &MockClient{}

			// 설정
			cache.Set(key, mockClient)

			// 조회
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

	// 통계 확인
	stats := cache.Stats()
	if stats["total_items"] != 10 {
		t.Errorf("Expected 10 total items, got %v", stats["total_items"])
	}
}

// MockClient 테스트용 모의 클라이언트
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
