package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/taking/kubemigrate/pkg/client"
)

// CachedClient 캐시된 클라이언트 정보
type CachedClient struct {
	Client    client.Client
	ApiType   string
	Config    map[string]interface{} // 원본 설정 정보 저장
	CreatedAt time.Time
	TTL       time.Duration
}

// IsExpired 캐시 만료 여부 확인
func (c *CachedClient) IsExpired() bool {
	return time.Since(c.CreatedAt) > c.TTL
}

// Manager 캐시 매니저
type Manager struct {
	clients     map[string]*CachedClient
	mutex       sync.RWMutex
	ttl         time.Duration
	hitCount    int64
	missCount   int64
	createCount int64
}

// NewManager 새로운 캐시 매니저 생성
func NewManager(ttl time.Duration) *Manager {
	return &Manager{
		clients: make(map[string]*CachedClient),
		ttl:     ttl,
	}
}

// GetCachedClient 캐시된 클라이언트 조회 또는 생성
func (m *Manager) GetCachedClient(apiType string, config map[string]interface{}) (client.Client, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 캐시 키 생성
	cacheKey := m.generateSimpleCacheKey(apiType, config)

	// 캐시에서 조회
	if cachedClient, exists := m.clients[cacheKey]; exists && !cachedClient.IsExpired() {
		m.hitCount++
		return cachedClient.Client, nil
	}

	// 캐시 미스
	m.missCount++

	// 캐시에 없거나 만료되었으면 새로 생성
	unifiedClient := client.NewClient()
	m.clients[cacheKey] = &CachedClient{
		Client:    unifiedClient,
		ApiType:   apiType,
		Config:    config, // 설정 정보 저장
		CreatedAt: time.Now(),
		TTL:       m.ttl,
	}

	m.createCount++
	return unifiedClient, nil
}

// generateSimpleCacheKey 설정 기반 캐시 키 생성
func (m *Manager) generateSimpleCacheKey(apiType string, config map[string]interface{}) string {
	// 설정을 정렬된 JSON으로 변환하여 일관된 해시 생성
	configBytes, err := json.Marshal(config)
	if err != nil {
		// JSON 변환 실패 시 기본 해시 생성
		configStr := fmt.Sprintf("api:%s|config:%v", apiType, config)
		hash := sha256.Sum256([]byte(configStr))
		return hex.EncodeToString(hash[:])
	}

	// API 타입과 설정을 조합하여 키 생성
	keyData := fmt.Sprintf("api:%s|config:%s", apiType, string(configBytes))
	hash := sha256.Sum256([]byte(keyData))
	return hex.EncodeToString(hash[:])
}

// generateReadableCacheKey 읽기 쉬운 캐시 키 생성 (디버깅용)
func (m *Manager) generateReadableCacheKey(apiType string, config map[string]interface{}) string {
	// 주요 설정만 추출하여 읽기 쉬운 키 생성
	keyParts := []string{apiType}

	if kubeConfig, ok := config["kubeconfig"].(string); ok && kubeConfig != "" {
		// kubeconfig의 마지막 8자리만 사용
		if len(kubeConfig) > 8 {
			keyParts = append(keyParts, "kube:"+kubeConfig[len(kubeConfig)-8:])
		} else {
			keyParts = append(keyParts, "kube:"+kubeConfig)
		}
	}

	if endpoint, ok := config["minio_endpoint"].(string); ok && endpoint != "" {
		keyParts = append(keyParts, "minio:"+endpoint)
	}

	if accessKey, ok := config["minio_access_key"].(string); ok && accessKey != "" {
		// access key의 첫 4자리만 사용
		if len(accessKey) > 4 {
			keyParts = append(keyParts, "key:"+accessKey[:4])
		} else {
			keyParts = append(keyParts, "key:"+accessKey)
		}
	}

	return strings.Join(keyParts, "|")
}

// GetStats 캐시 통계 조회
func (m *Manager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 만료된 캐시 정리
	expiredCount := 0
	for _, cachedClient := range m.clients {
		if cachedClient.IsExpired() {
			expiredCount++
		}
	}

	// 히트율 계산
	totalRequests := m.hitCount + m.missCount
	var hitRate, missRate float64
	if totalRequests > 0 {
		hitRate = float64(m.hitCount) / float64(totalRequests) * 100
		missRate = float64(m.missCount) / float64(totalRequests) * 100
	}

	// 활성 클라이언트 상세 정보 수집
	activeClients := make([]map[string]interface{}, 0)

	for key, cachedClient := range m.clients {
		// 만료된 클라이언트는 건너뛰기
		if cachedClient.IsExpired() {
			continue
		}

		// 저장된 설정 정보를 사용하여 읽기 쉬운 키 생성
		readableKey := m.generateReadableCacheKey(cachedClient.ApiType, cachedClient.Config)

		clientInfo := map[string]interface{}{
			"cache_key":         key,
			"readable_key":      readableKey,
			"api_type":          cachedClient.ApiType,
			"config":            cachedClient.Config,
			"created_at":        cachedClient.CreatedAt.Format("2006-01-02 15:04:05"),
			"age_seconds":       int(time.Since(cachedClient.CreatedAt).Seconds()),
			"ttl_seconds":       int(cachedClient.TTL.Seconds()),
			"is_expired":        cachedClient.IsExpired(),
			"expires_at":        cachedClient.CreatedAt.Add(cachedClient.TTL).Format("2006-01-02 15:04:05"),
			"remaining_seconds": int(cachedClient.TTL.Seconds() - time.Since(cachedClient.CreatedAt).Seconds()),
		}

		activeClients = append(activeClients, clientInfo)
	}

	// 활성 클라이언트만 간소화된 정보로 표시
	simplifiedActiveClients := make([]map[string]interface{}, 0)
	for _, client := range activeClients {
		simplifiedClient := map[string]interface{}{
			"api_type":          client["api_type"],
			"readable_key":      client["readable_key"],
			"age_seconds":       client["age_seconds"],
			"remaining_seconds": client["remaining_seconds"],
		}

		// 설정 정보가 있으면 주요 설정만 표시
		if config, ok := client["config"].(map[string]interface{}); ok && len(config) > 0 {
			simplifiedConfig := make(map[string]interface{})
			for key, value := range config {
				// 민감한 정보는 마스킹
				if key == "kubeconfig" || key == "minio_secret_key" {
					if str, ok := value.(string); ok && len(str) > 8 {
						simplifiedConfig[key] = "***" + str[len(str)-8:]
					} else {
						simplifiedConfig[key] = "***"
					}
				} else {
					simplifiedConfig[key] = value
				}
			}
			simplifiedClient["config"] = simplifiedConfig
		}

		simplifiedActiveClients = append(simplifiedActiveClients, simplifiedClient)
	}

	// 순서가 보장되는 map 생성 (Go 1.18+에서는 map 순서가 보장됨)
	stats := make(map[string]interface{})

	// 1. summary 먼저
	stats["summary"] = map[string]interface{}{
		"total_clients":   len(m.clients),
		"active_clients":  len(m.clients) - expiredCount,
		"expired_clients": expiredCount,
	}

	// 2. active_clients
	stats["active_clients"] = simplifiedActiveClients

	// 3. performance 마지막
	stats["performance"] = map[string]interface{}{
		"hit_rate":       fmt.Sprintf("%.2f%%", hitRate),
		"miss_rate":      fmt.Sprintf("%.2f%%", missRate),
		"total_requests": m.hitCount + m.missCount,
	}

	return stats
}

// Cleanup 캐시 정리
func (m *Manager) Cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 만료된 캐시 제거
	for key, cachedClient := range m.clients {
		if cachedClient.IsExpired() {
			delete(m.clients, key)
		}
	}
}

// Invalidate 특정 설정의 캐시 무효화
func (m *Manager) Invalidate(apiType string, config map[string]interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 캐시 키 생성
	cacheKey := m.generateSimpleCacheKey(apiType, config)

	// 특정 캐시 무효화
	delete(m.clients, cacheKey)
}

// InvalidateAll 모든 캐시 무효화
func (m *Manager) InvalidateAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 모든 캐시 무효화
	m.clients = make(map[string]*CachedClient)
}

// GetCacheKey 캐시 키 생성 (디버깅용)
func (m *Manager) GetCacheKey(apiType string, config map[string]interface{}) string {
	return m.generateSimpleCacheKey(apiType, config)
}

// GenerateCacheKey 공개 캐시 키 생성 메서드
func (m *Manager) GenerateCacheKey(apiType string, config map[string]interface{}) string {
	return m.generateSimpleCacheKey(apiType, config)
}

// GetCacheInfo 특정 설정의 캐시 정보 조회
func (m *Manager) GetCacheInfo(apiType string, config map[string]interface{}) map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	cacheKey := m.GetCacheKey(apiType, config)

	// 캐시에서 조회 시도
	cachedClient, exists := m.clients[cacheKey]
	exists = exists && !cachedClient.IsExpired()

	return map[string]interface{}{
		"api_type":  apiType,
		"cache_key": cacheKey,
		"config":    config,
		"exists":    exists,
		"status":    "cache_info_retrieved",
	}
}
