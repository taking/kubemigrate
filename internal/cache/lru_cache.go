package cache

import (
	"container/list"
	"strings"
	"sync"
	"time"

	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
	"github.com/taking/kubemigrate/pkg/utils"
)

// LRUItem : LRU 캐시의 항목
type LRUItem struct {
	Key          string
	Value        client.Client
	CreatedAt    time.Time
	LastAccess   time.Time
	ApiType      string
	KubeConfig   config.KubeConfig
	VeleroConfig config.VeleroConfig
	MinioConfig  config.MinioConfig
}

// LRUCache : 메모리 효율적인 LRU 캐시
type LRUCache struct {
	capacity    int
	items       map[string]*list.Element
	list        *list.List
	mutex       sync.RWMutex
	totalHits   int64
	totalMisses int64
}

// NewLRUCache : 새로운 LRU 캐시 생성
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Get : 캐시에서 값 조회
func (c *LRUCache) Get(key string) (client.Client, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, exists := c.items[key]; exists {
		item := elem.Value.(*LRUItem)

		item.LastAccess = time.Now()
		c.list.MoveToFront(elem)
		c.totalHits++
		return item.Value, true
	}

	c.totalMisses++
	return nil, false
}

// Set : 캐시에 값 저장
func (c *LRUCache) Set(key string, value client.Client) {
	c.SetWithApiType(key, value, "unknown")
}

// SetWithApiType : API 타입을 명시적으로 지정하여 캐시에 값 저장
func (c *LRUCache) SetWithApiType(key string, value client.Client, apiType string) {
	c.SetWithConfigs(key, value, apiType, config.KubeConfig{}, config.VeleroConfig{}, config.MinioConfig{})
}

// SetWithConfigs : 설정 정보와 함께 캐시에 값 저장
func (c *LRUCache) SetWithConfigs(key string, value client.Client, apiType string, kubeConfig config.KubeConfig, veleroConfig config.VeleroConfig, minioConfig config.MinioConfig) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, exists := c.items[key]; exists {
		item := elem.Value.(*LRUItem)
		item.Value = value
		item.LastAccess = time.Now()
		item.ApiType = apiType
		item.KubeConfig = kubeConfig
		item.VeleroConfig = veleroConfig
		item.MinioConfig = minioConfig
		c.list.MoveToFront(elem)
		return
	}

	item := &LRUItem{
		Key:          key,
		Value:        value,
		CreatedAt:    time.Now(),
		LastAccess:   time.Now(),
		ApiType:      apiType,
		KubeConfig:   kubeConfig,
		VeleroConfig: veleroConfig,
		MinioConfig:  minioConfig,
	}

	// 캐시가 가득 찬 경우 오래된 항목 제거
	if c.list.Len() >= c.capacity {
		c.evictOldest()
	}

	// 새 항목을 리스트 맨 앞에 추가
	elem := c.list.PushFront(item)
	c.items[key] = elem
}

// Remove : 캐시에서 항목 제거
func (c *LRUCache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, exists := c.items[key]; exists {
		c.removeElement(elem)
	}
}

// Cleanup : LRU 전체 캐시 제거
func (c *LRUCache) Cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*list.Element)
	c.list = list.New()
}

// CleanByKey : 특정 키의 캐시 항목 제거
func (c *LRUCache) CleanByKey(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, exists := c.items[key]; exists {
		c.removeElement(elem)
		return true
	}
	return false
}

// CleanByPattern : 패턴에 맞는 캐시 항목 제거
func (c *LRUCache) CleanByPattern(pattern string) int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	removedCount := 0
	keysToRemove := make([]string, 0)

	// 제거할 키들을 먼저 수집
	for key := range c.items {
		if strings.Contains(key, pattern) {
			keysToRemove = append(keysToRemove, key)
		}
	}

	// 수집된 키들을 제거
	for _, key := range keysToRemove {
		if elem, exists := c.items[key]; exists {
			c.removeElement(elem)
			removedCount++
		}
	}

	return removedCount
}

// Stats : 캐시 통계 반환
func (c *LRUCache) Stats() map[string]interface{} {
	detailedStats := c.GetDetailedStats()

	return map[string]interface{}{
		"total_items":    detailedStats.Summary.TotalClients,
		"active_items":   detailedStats.Summary.ActiveClients,
		"capacity":       detailedStats.Summary.Capacity,
		"active_clients": detailedStats.ActiveClients,
	}
}

// GetDetailedStats : 상세한 캐시 통계 반환
func (c *LRUCache) GetDetailedStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	activeClients := make([]ActiveClientInfo, 0)
	var totalAge int64
	var oldestAge, newestAge int

	for _, elem := range c.items {
		item := elem.Value.(*LRUItem)
		ageSeconds := calculateAgeSeconds(item.CreatedAt)

		clientInfo := ActiveClientInfo{
			ApiType:     item.ApiType,
			CacheKey:    item.Key,
			ReadableKey: generateReadableKey(item.ApiType, item.Key),
			CreatedAt:   item.CreatedAt,
			AgeSeconds:  ageSeconds,
			Config:      c.getMaskedConfig(item.ApiType, item.Key),
		}

		// LRU 캐시에서는 모든 항목이 활성 상태
		activeClients = append(activeClients, clientInfo)
		totalAge += int64(ageSeconds)

		if oldestAge == 0 || ageSeconds > oldestAge {
			oldestAge = ageSeconds
		}
		if newestAge == 0 || ageSeconds < newestAge {
			newestAge = ageSeconds
		}
	}

	// 성능 통계 계산
	totalRequests := c.totalHits + c.totalMisses
	var hitRate, missRate float64
	if totalRequests > 0 {
		hitRate = float64(c.totalHits) / float64(totalRequests) * 100
		missRate = float64(c.totalMisses) / float64(totalRequests) * 100
	}

	var averageAge float64
	if len(activeClients) > 0 {
		averageAge = float64(totalAge) / float64(len(activeClients))
	}

	return CacheStats{
		Summary: CacheSummary{
			ActiveClients: len(activeClients),
			TotalClients:  len(c.items),
			Capacity:      c.capacity,
		},
		ActiveClients: activeClients,
		Performance: PerformanceStats{
			HitRate:      hitRate,
			MissRate:     missRate,
			TotalHits:    c.totalHits,
			TotalMisses:  c.totalMisses,
			AverageAge:   averageAge,
			OldestClient: oldestAge,
			NewestClient: newestAge,
		},
	}
}

// GetOrCreate : 캐시에서 조회하거나 새로 생성
func (c *LRUCache) GetOrCreate(
	kubeConfig config.KubeConfig,
	helmConfig config.KubeConfig,
	veleroConfig config.VeleroConfig,
	minioConfig config.MinioConfig,
	createFunc func() client.Client,
) client.Client {
	key := utils.GenerateCompositeCacheKey(
		kubeConfig.KubeConfig,
		helmConfig.KubeConfig,
		veleroConfig.KubeConfig.KubeConfig,
		minioConfig.Endpoint,
		minioConfig.AccessKey,
		minioConfig.SecretKey,
	)

	// 캐시에서 조회 시도
	if cached, exists := c.Get(key); exists {
		return cached
	}

	// 캐시에 없으면 새로 생성
	newClient := createFunc()

	// API 타입 결정 (설정이 있는 서비스 우선)
	apiType := c.determineApiType(kubeConfig, helmConfig, veleroConfig, minioConfig)
	c.SetWithApiType(key, newClient, apiType)

	return newClient
}

// GetOrCreateWithApiType : API 타입을 명시적으로 지정하여 캐시에서 조회하거나 새로 생성
func (c *LRUCache) GetOrCreateWithApiType(
	kubeConfig config.KubeConfig,
	helmConfig config.KubeConfig,
	veleroConfig config.VeleroConfig,
	minioConfig config.MinioConfig,
	apiType string,
	createFunc func() client.Client,
) client.Client {
	// 복합 캐시 키 생성
	key := utils.GenerateCompositeCacheKey(
		kubeConfig.KubeConfig,
		helmConfig.KubeConfig,
		veleroConfig.KubeConfig.KubeConfig,
		minioConfig.Endpoint,
		minioConfig.AccessKey,
		minioConfig.SecretKey,
		apiType,
	)

	// 캐시에서 조회 시도
	if cached, exists := c.Get(key); exists {
		return cached
	}

	// 캐시에 없으면 새로 생성
	newClient := createFunc()
	c.SetWithConfigs(key, newClient, apiType, kubeConfig, veleroConfig, minioConfig)

	return newClient
}

// determineApiType : 설정을 기반으로 API 타입을 결정
func (c *LRUCache) determineApiType(
	kubeConfig config.KubeConfig,
	helmConfig config.KubeConfig,
	veleroConfig config.VeleroConfig,
	minioConfig config.MinioConfig,
) string {
	// 설정이 있는 서비스를 우선적으로 선택
	if len(minioConfig.Endpoint) > 0 {
		return "minio"
	} else if len(veleroConfig.KubeConfig.KubeConfig) > 0 {
		return "velero"
	} else if len(helmConfig.KubeConfig) > 0 {
		return "helm"
	} else if len(kubeConfig.KubeConfig) > 0 {
		return "kubernetes"
	}

	// 기본값
	return "kubernetes"
}

// evictOldest : 가장 오래된 항목 제거
func (c *LRUCache) evictOldest() {
	if c.list.Len() == 0 {
		return
	}

	// 리스트의 마지막 요소(가장 오래된) 제거
	elem := c.list.Back()
	c.removeElement(elem)
}

// removeElement : 리스트에서 요소 제거
func (c *LRUCache) removeElement(elem *list.Element) {
	item := elem.Value.(*LRUItem)
	c.list.Remove(elem)
	delete(c.items, item.Key)
}

// detectApiType : 클라이언트 타입을 분석하여 API 타입을 결정
func (c *LRUCache) detectApiType(key string, client client.Client) string {
	return getApiTypeFromKey(key)
}

// getMaskedConfig : API 타입에 따라 마스킹된 설정 반환
func (c *LRUCache) getMaskedConfig(apiType, key string) any {
	// 캐시에서 해당 키의 설정 정보 찾기
	if elem, exists := c.items[key]; exists {
		item := elem.Value.(*LRUItem)

		switch apiType {
		case "kubernetes":
			return maskKubernetesConfig(item.KubeConfig)
		case "minio":
			return maskMinioConfig(item.MinioConfig)
		case "velero":
			return maskVeleroConfig(item.VeleroConfig, item.MinioConfig)
		case "helm":
			return maskHelmConfig(item.KubeConfig)
		default:
			return map[string]interface{}{
				"api_type": apiType,
				"key":      maskString(key),
			}
		}
	}

	// 캐시에 없는 경우 기본값 반환
	switch apiType {
	case "kubernetes":
		return MaskedKubeConfig{
			KubeConfig: maskString(key),
			HasConfig:  false,
		}
	case "minio":
		return MaskedMinioConfig{
			Endpoint:  "",
			AccessKey: "",
			SecretKey: "",
			UseSSL:    false,
			HasConfig: false,
		}
	case "velero":
		return MaskedVeleroConfig{
			Kubernetes: MaskedKubeConfig{HasConfig: false},
			Minio:      MaskedMinioConfig{HasConfig: false},
			HasConfig:  false,
		}
	case "helm":
		return MaskedHelmConfig{
			KubeConfig: maskString(key),
			HasConfig:  false,
		}
	default:
		return map[string]interface{}{
			"api_type": apiType,
			"key":      maskString(key),
		}
	}
}
