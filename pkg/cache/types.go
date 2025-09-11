package cache

import (
	"time"

	"github.com/taking/kubemigrate/pkg/client"
)

// CacheConfig 캐시 설정
type CacheConfig struct {
	TTL time.Duration `json:"ttl"`
}

// CacheStats 캐시 통계 정보
type CacheStats struct {
	Summary       map[string]interface{}   `json:"summary"`
	ActiveClients []map[string]interface{} `json:"active_clients"`
	Performance   map[string]interface{}   `json:"performance"`
}

// CacheInfo 캐시 정보
type CacheInfo struct {
	ApiType  string                 `json:"api_type"`
	CacheKey string                 `json:"cache_key"`
	Config   map[string]interface{} `json:"config"`
	Exists   bool                   `json:"exists"`
	Status   string                 `json:"status"`
}

// ClientFactory 클라이언트 팩토리 인터페이스
type ClientFactory interface {
	CreateClient(apiType string, config map[string]interface{}) (client.Client, error)
}

// DefaultClientFactory 기본 클라이언트 팩토리
type DefaultClientFactory struct{}

// CreateClient 기본 클라이언트 생성
func (f *DefaultClientFactory) CreateClient(apiType string, config map[string]interface{}) (client.Client, error) {
	return client.NewClient(), nil
}
