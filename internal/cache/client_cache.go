package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/pkg/client"
)

// CachedClient : 캐시된 클라이언트 정보
type CachedClient struct {
	Client    client.Client
	ApiType   string
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

// generateCacheKey : 설정 조합으로 캐시 키 생성
func (cc *ClientCache) generateCacheKey(apiType string, kubeConfig config.KubeConfig, helmConfig config.KubeConfig, veleroConfig config.VeleroConfig, minioConfig config.MinioConfig) string {
	// API 타입과 모든 설정을 조합하여 키 생성
	configStr := fmt.Sprintf("api:%s|kube:%s|helm:%s|velero:%s|minio:%s:%s:%s:%t",
		apiType,
		kubeConfig.Config,
		helmConfig.Config,
		veleroConfig.KubeConfig.Config,
		minioConfig.Endpoint,
		minioConfig.AccessKey,
		minioConfig.SecretKey,
		minioConfig.UseSSL)

	hash := sha256.Sum256([]byte(configStr))
	return hex.EncodeToString(hash[:])
}

// Get : 캐시에서 클라이언트 조회
func (cc *ClientCache) Get(apiType string, kubeConfig config.KubeConfig, helmConfig config.KubeConfig, veleroConfig config.VeleroConfig, minioConfig config.MinioConfig) (client.Client, bool) {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	key := cc.generateCacheKey(apiType, kubeConfig, helmConfig, veleroConfig, minioConfig)
	cached, exists := cc.clients[key]

	if !exists || cached.IsExpired() {
		return nil, false
	}

	return cached.Client, true
}

// Set : 캐시에 클라이언트 저장
func (cc *ClientCache) Set(apiType string, kubeConfig config.KubeConfig, helmConfig config.KubeConfig, veleroConfig config.VeleroConfig, minioConfig config.MinioConfig, client client.Client) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	key := cc.generateCacheKey(apiType, kubeConfig, helmConfig, veleroConfig, minioConfig)
	cc.clients[key] = &CachedClient{
		Client:    client,
		ApiType:   apiType,
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
	apiType string,
	kubeConfig config.KubeConfig,
	helmConfig config.KubeConfig,
	veleroConfig config.VeleroConfig,
	minioConfig config.MinioConfig,
	createFunc func() client.Client,
) client.Client {
	// 캐시에서 조회 시도
	if cached, exists := cc.Get(apiType, kubeConfig, helmConfig, veleroConfig, minioConfig); exists {
		return cached
	}

	// 캐시에 없으면 새로 생성
	newClient := createFunc()
	cc.Set(apiType, kubeConfig, helmConfig, veleroConfig, minioConfig, newClient)

	return newClient
}

// Invalidate : 특정 설정의 캐시 무효화
func (cc *ClientCache) Invalidate(apiType string, kubeConfig config.KubeConfig, helmConfig config.KubeConfig, veleroConfig config.VeleroConfig, minioConfig config.MinioConfig) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	key := cc.generateCacheKey(apiType, kubeConfig, helmConfig, veleroConfig, minioConfig)
	delete(cc.clients, key)
}

// InvalidateAll : 모든 캐시 무효화
func (cc *ClientCache) InvalidateAll() {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	cc.clients = make(map[string]*CachedClient)
}

// Stats : 캐시 통계 정보
func (cc *ClientCache) Stats() map[string]interface{} {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	active := 0
	expired := 0
	var activeClients []map[string]interface{}
	var expiredClients []map[string]interface{}

	// API 타입별 통계
	apiTypeStats := make(map[string]int)

	for key, cached := range cc.clients {
		apiType := cached.ApiType
		clientInfo := map[string]interface{}{
			"api_type":    apiType,
			"cache_key":   key,
			"created_at":  cached.CreatedAt.Format("2006-01-02 15:04:05"),
			"ttl":         cached.TTL.String(),
			"expires_at":  cached.CreatedAt.Add(cached.TTL).Format("2006-01-02 15:04:05"),
			"is_expired":  cached.IsExpired(),
			"client_type": cc.getClientType(cached.Client),
		}

		// API 타입별 통계 증가
		apiTypeStats[apiType]++

		if cached.IsExpired() {
			expired++
			expiredClients = append(expiredClients, clientInfo)
		} else {
			active++
			activeClients = append(activeClients, clientInfo)
		}
	}

	return map[string]interface{}{
		"total_clients":   len(cc.clients),
		"active_clients":  active,
		"expired_clients": expired,
		"cache_ttl":       cc.ttl.String(),
		"api_type_stats":  apiTypeStats,
		"active_details":  activeClients,
		"expired_details": expiredClients,
	}
}

// getClientType : 클라이언트 타입 식별
func (cc *ClientCache) getClientType(client client.Client) string {
	if client == nil {
		return "unknown"
	}

	// 리플렉션을 사용하여 클라이언트 타입 확인
	clientType := reflect.TypeOf(client).String()

	// 패키지 경로에서 클라이언트 타입 추출
	if strings.Contains(clientType, "kubernetes") {
		return "kubernetes"
	} else if strings.Contains(clientType, "helm") {
		return "helm"
	} else if strings.Contains(clientType, "velero") {
		return "velero"
	} else if strings.Contains(clientType, "minio") {
		return "minio"
	}

	return "unified"
}
