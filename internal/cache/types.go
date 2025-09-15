// Package cache 메모리 효율적인 LRU 캐시를 제공합니다.
package cache

import (
	"time"
)

// CacheStats : 캐시 통계 정보
type CacheStats struct {
	Summary       CacheSummary       `json:"summary"`
	ActiveClients []ActiveClientInfo `json:"active_clients"`
	Performance   PerformanceStats   `json:"performance"`
}

// CacheSummary : 캐시 요약 정보
type CacheSummary struct {
	ActiveClients int `json:"active_clients"`
	TotalClients  int `json:"total_clients"`
	Capacity      int `json:"capacity"`
}

// ActiveClientInfo : 활성 클라이언트 정보
type ActiveClientInfo struct {
	ApiType     string    `json:"api_type"`
	CacheKey    string    `json:"cache_key"`
	ReadableKey string    `json:"readable_key"`
	CreatedAt   time.Time `json:"created_at"`
	AgeSeconds  int       `json:"age_seconds"`
	Config      any       `json:"config"` // 마스킹된 설정 정보
}

// PerformanceStats : 성능 통계
type PerformanceStats struct {
	HitRate      float64 `json:"hit_rate"`
	MissRate     float64 `json:"miss_rate"`
	TotalHits    int64   `json:"total_hits"`
	TotalMisses  int64   `json:"total_misses"`
	AverageAge   float64 `json:"average_age_seconds"`
	OldestClient int     `json:"oldest_client_seconds"`
	NewestClient int     `json:"newest_client_seconds"`
}

// MaskedKubeConfig : 마스킹된 Kubernetes 설정
type MaskedKubeConfig struct {
	KubeConfig string `json:"kubeconfig"` // 마스킹된 kubeconfig
	HasConfig  bool   `json:"has_config"`
}

// MaskedKubernetesConfig : 마스킹된 Kubernetes 설정 (별칭)
type MaskedKubernetesConfig = MaskedKubeConfig

// MaskedMinioConfig : 마스킹된 MinIO 설정
type MaskedMinioConfig struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"` // 마스킹된 access key
	SecretKey string `json:"secret_key"` // 마스킹된 secret key
	UseSSL    bool   `json:"use_ssl"`
	HasConfig bool   `json:"has_config"`
}

// MaskedVeleroConfig : 마스킹된 Velero 설정
type MaskedVeleroConfig struct {
	Kubernetes MaskedKubernetesConfig `json:"kubernetes"`
	Minio      MaskedMinioConfig      `json:"minio"`
	HasConfig  bool                   `json:"has_config"`
}

// MaskedHelmConfig : 마스킹된 Helm 설정
type MaskedHelmConfig struct {
	KubeConfig string `json:"kubeconfig"` // 마스킹된 kubeconfig
	HasConfig  bool   `json:"has_config"`
}
