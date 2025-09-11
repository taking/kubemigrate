package cache

import (
	"time"

	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/pkg/client"
)

// CacheConfig 캐시 설정
type CacheConfig struct {
	TTL time.Duration `json:"ttl"`
}

// CacheStats 캐시 통계 정보
type CacheStats struct {
	Summary       CacheSummary       `json:"summary"`
	ActiveClients []ActiveClientInfo `json:"active_clients"`
	Performance   PerformanceStats   `json:"performance"`
}

// CacheSummary 캐시 요약 정보
type CacheSummary struct {
	ActiveClients  int `json:"active_clients"`
	ExpiredClients int `json:"expired_clients"`
	TotalClients   int `json:"total_clients"`
}

// ActiveClientInfo 활성 클라이언트 정보
type ActiveClientInfo struct {
	ApiType          string    `json:"api_type"`
	CacheKey         string    `json:"cache_key"`
	ReadableKey      string    `json:"readable_key"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	AgeSeconds       int       `json:"age_seconds"`
	RemainingSeconds int       `json:"remaining_seconds"`
	TTLSeconds       int       `json:"ttl_seconds"`
	IsExpired        bool      `json:"is_expired"`
	Config           any       `json:"config"` // 구체적인 타입으로 변경
}

// PerformanceStats 성능 통계
type PerformanceStats struct {
	HitCount      int    `json:"hit_count"`
	MissCount     int    `json:"miss_count"`
	CreateCount   int    `json:"create_count"`
	HitRate       string `json:"hit_rate"`
	MissRate      string `json:"miss_rate"`
	TotalRequests int    `json:"total_requests"`
}

// CacheInfo 캐시 정보
type CacheInfo struct {
	ApiType  string `json:"api_type"`
	CacheKey string `json:"cache_key"`
	Config   any    `json:"config"` // 구체적인 타입으로 변경
	Exists   bool   `json:"exists"`
	Status   string `json:"status"`
}

// ClientFactory 클라이언트 팩토리 인터페이스
type ClientFactory interface {
	CreateClient(apiType string, config map[string]interface{}) (client.Client, error)
	CreateClientWithTypedConfig(apiType string, config *config.PluginConfigData) (client.Client, error)
}

// DefaultClientFactory 기본 클라이언트 팩토리
type DefaultClientFactory struct{}

// CreateClient 기본 클라이언트 생성 (기존 호환성 유지)
func (f *DefaultClientFactory) CreateClient(apiType string, config map[string]interface{}) (client.Client, error) {
	return client.NewClient(), nil
}

// CreateClientWithTypedConfig 타입 안전 설정으로 클라이언트 생성
func (f *DefaultClientFactory) CreateClientWithTypedConfig(apiType string, config *config.PluginConfigData) (client.Client, error) {
	return client.NewClient(), nil
}
