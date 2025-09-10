package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/pkg/client"
)

// CachedClient : 캐시된 클라이언트 정보
type CachedClient struct {
	Client    client.Client
	CreatedAt time.Time
	TTL       time.Duration
}

// IsExpired : 캐시 만료 여부 확인
func (c *CachedClient) IsExpired() bool {
	return time.Since(c.CreatedAt) > c.TTL
}

// ClientCache : 클라이언트 캐시 관리자
type ClientCache struct {
	clients map[string]*CachedClient
	mutex   sync.RWMutex
	ttl     time.Duration
}

// NewClientCache : 새로운 클라이언트 캐시 생성
func NewClientCache(ttl time.Duration) *ClientCache {
	return &ClientCache{
		clients: make(map[string]*CachedClient),
		ttl:     ttl,
	}
}

// generateCacheKey : kubeconfig 해시로 캐시 키 생성
func (cc *ClientCache) generateCacheKey(kubeConfig string) string {
	hash := sha256.Sum256([]byte(kubeConfig))
	return hex.EncodeToString(hash[:])
}

// Get : 캐시에서 클라이언트 조회
func (cc *ClientCache) Get(kubeConfig string) (client.Client, bool) {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	key := cc.generateCacheKey(kubeConfig)
	cached, exists := cc.clients[key]

	if !exists || cached.IsExpired() {
		return nil, false
	}

	return cached.Client, true
}

// Set : 캐시에 클라이언트 저장
func (cc *ClientCache) Set(kubeConfig string, client client.Client) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	key := cc.generateCacheKey(kubeConfig)
	cc.clients[key] = &CachedClient{
		Client:    client,
		CreatedAt: time.Now(),
		TTL:       cc.ttl,
	}
}

// Cleanup : 만료된 캐시 정리
func (cc *ClientCache) Cleanup() {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	for key, cached := range cc.clients {
		if cached.IsExpired() {
			delete(cc.clients, key)
		}
	}
}

// GetOrCreate : 캐시에서 조회하거나 새로 생성
func (cc *ClientCache) GetOrCreate(
	kubeConfig config.KubeConfig,
	helmConfig config.KubeConfig,
	veleroConfig config.VeleroConfig,
	minioConfig config.MinioConfig,
	createFunc func() client.Client,
) client.Client {
	// 캐시에서 조회 시도
	if cached, exists := cc.Get(kubeConfig.KubeConfig); exists {
		return cached
	}

	// 캐시에 없으면 새로 생성
	newClient := createFunc()
	cc.Set(kubeConfig.KubeConfig, newClient)

	return newClient
}

// Stats : 캐시 통계 정보
func (cc *ClientCache) Stats() map[string]interface{} {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	active := 0
	expired := 0

	for _, cached := range cc.clients {
		if cached.IsExpired() {
			expired++
		} else {
			active++
		}
	}

	return map[string]interface{}{
		"total_clients":   len(cc.clients),
		"active_clients":  active,
		"expired_clients": expired,
		"cache_ttl":       cc.ttl.String(),
	}
}
