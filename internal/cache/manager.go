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
	Config    any // 구체적인 타입으로 변경
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
func (m *Manager) generateReadableCacheKey(apiType string, config any) string {
	// 주요 설정만 추출하여 읽기 쉬운 키 생성
	keyParts := []string{apiType}

	// config가 map[string]interface{}인 경우 처리
	if configMap, ok := config.(map[string]interface{}); ok {
		if kubeConfig, ok := configMap["kubeconfig"].(string); ok && kubeConfig != "" {
			// kubeconfig의 마지막 8자리만 사용
			if len(kubeConfig) > 8 {
				keyParts = append(keyParts, "kube:"+kubeConfig[len(kubeConfig)-8:])
			} else {
				keyParts = append(keyParts, "kube:"+kubeConfig)
			}
		}

		if endpoint, ok := configMap["minio_endpoint"].(string); ok && endpoint != "" {
			keyParts = append(keyParts, "minio:"+endpoint)
		}

		if accessKey, ok := configMap["minio_access_key"].(string); ok && accessKey != "" {
			// access key의 첫 4자리만 사용
			if len(accessKey) > 4 {
				keyParts = append(keyParts, "key:"+accessKey[:4])
			} else {
				keyParts = append(keyParts, "key:"+accessKey)
			}
		}
	}

	return strings.Join(keyParts, "|")
}

// GetStats 캐시 통계 조회
func (m *Manager) GetStats() CacheStats {
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
	activeClients := make([]ActiveClientInfo, 0)

	for key, cachedClient := range m.clients {
		// 만료된 클라이언트는 건너뛰기
		if cachedClient.IsExpired() {
			continue
		}

		// 저장된 설정 정보를 사용하여 읽기 쉬운 키 생성
		readableKey := m.generateReadableCacheKey(cachedClient.ApiType, cachedClient.Config)

		clientInfo := ActiveClientInfo{
			ApiType:          cachedClient.ApiType,
			CacheKey:         key,
			ReadableKey:      readableKey,
			CreatedAt:        cachedClient.CreatedAt,
			ExpiresAt:        cachedClient.CreatedAt.Add(cachedClient.TTL),
			AgeSeconds:       int(time.Since(cachedClient.CreatedAt).Seconds()),
			RemainingSeconds: int(cachedClient.TTL.Seconds() - time.Since(cachedClient.CreatedAt).Seconds()),
			TTLSeconds:       int(cachedClient.TTL.Seconds()),
			IsExpired:        cachedClient.IsExpired(),
			Config:           maskSensitiveConfig(cachedClient.Config),
		}

		activeClients = append(activeClients, clientInfo)
	}

	// 통계 정보 구성
	stats := CacheStats{
		Summary: CacheSummary{
			TotalClients:   len(m.clients),
			ActiveClients:  len(m.clients) - expiredCount,
			ExpiredClients: expiredCount,
		},
		ActiveClients: activeClients,
		Performance: PerformanceStats{
			HitCount:      int(m.hitCount),
			MissCount:     int(m.missCount),
			CreateCount:   int(m.createCount),
			HitRate:       fmt.Sprintf("%.2f%%", hitRate),
			MissRate:      fmt.Sprintf("%.2f%%", missRate),
			TotalRequests: int(totalRequests),
		},
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
		"config":    maskSensitiveConfig(config),
		"exists":    exists,
		"status":    "cache_info_retrieved",
	}
}

// maskSensitiveConfig 민감한 정보를 마스킹 처리
func maskSensitiveConfig(config any) any {
	if config == nil {
		return nil
	}

	// map[string]interface{} 타입인 경우
	if configMap, ok := config.(map[string]interface{}); ok {
		maskedConfig := make(map[string]interface{})

		for key, value := range configMap {
			switch key {
			case "kubeconfig":
				if str, ok := value.(string); ok && str != "" {
					// kubeconfig는 길이에 따라 마스킹
					if len(str) > 20 {
						maskedConfig[key] = str[:10] + "..." + str[len(str)-10:]
					} else {
						maskedConfig[key] = "***masked***"
					}
				} else {
					maskedConfig[key] = value
				}
			case "minio_access_key", "accessKey":
				if str, ok := value.(string); ok && str != "" {
					// access key는 앞 4자리만 보여주고 나머지 마스킹
					if len(str) > 4 {
						maskedConfig[key] = str[:4] + "***masked***"
					} else {
						maskedConfig[key] = "***masked***"
					}
				} else {
					maskedConfig[key] = value
				}
			case "minio_secret_key", "secretKey":
				// secret key는 완전 마스킹
				maskedConfig[key] = "***masked***"
			case "minio_endpoint", "endpoint":
				// endpoint는 그대로 유지 (민감하지 않음)
				maskedConfig[key] = value
			case "minio_use_ssl", "useSSL":
				// useSSL은 그대로 유지 (민감하지 않음)
				maskedConfig[key] = value
			case "namespace":
				// namespace는 그대로 유지 (민감하지 않음)
				maskedConfig[key] = value
			default:
				// 기타 필드는 그대로 유지
				maskedConfig[key] = value
			}
		}
		return maskedConfig
	}

	// 기타 타입은 그대로 반환
	return config
}
