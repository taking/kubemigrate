package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/pkg/client"
)

// CachePlugin 캐시 관리 플러그인
type CachePlugin struct {
	cacheManager *Manager
	config       map[string]interface{}
}

// NewPlugin 새로운 캐시 플러그인 생성
func NewPlugin() *CachePlugin {
	return &CachePlugin{
		cacheManager: nil, // 플러그인 매니저에서 설정됨
	}
}

// Name 플러그인 이름
func (p *CachePlugin) Name() string {
	return "cache"
}

// Version 플러그인 버전
func (p *CachePlugin) Version() string {
	return "1.0.0"
}

// Description 플러그인 설명
func (p *CachePlugin) Description() string {
	return "Client cache management plugin"
}

// Initialize 플러그인 초기화
func (p *CachePlugin) Initialize(config map[string]interface{}) error {
	p.config = config
	return nil
}

// Shutdown 플러그인 종료
func (p *CachePlugin) Shutdown() error {
	if p.cacheManager != nil {
		p.cacheManager.Cleanup()
	}
	return nil
}

// RegisterRoutes 라우트 등록
func (p *CachePlugin) RegisterRoutes(router *echo.Group) error {
	// 캐시 관련 라우트 등록
	cacheGroup := router.Group("/cache")

	// 캐시 통계 조회
	cacheGroup.GET("/stats", p.GetStatsHandler)

	// 캐시 정리
	cacheGroup.POST("/cleanup", p.CleanupHandler)

	// 캐시 무효화
	cacheGroup.POST("/invalidate", p.InvalidateHandler)

	// 모든 캐시 무효화
	cacheGroup.POST("/invalidate-all", p.InvalidateAllHandler)

	// 캐시 정보 조회
	cacheGroup.POST("/info", p.GetCacheInfoHandler)

	// 캐시 키 생성 (디버깅용)
	cacheGroup.POST("/key", p.GetCacheKeyHandler)

	return nil
}

// HealthCheck 헬스체크
func (p *CachePlugin) HealthCheck(ctx context.Context) error {
	// 캐시가 정상적으로 작동하는지 확인
	stats := p.cacheManager.GetStats()
	if stats == nil {
		return echo.NewHTTPError(500, "Cache not initialized")
	}
	return nil
}

// GetServiceType 서비스 타입
func (p *CachePlugin) GetServiceType() string {
	return "cache"
}

// GetClient 클라이언트 반환
func (p *CachePlugin) GetClient() interface{} {
	return p.cacheManager
}

// SetPluginManager 플러그인 매니저 설정
func (p *CachePlugin) SetPluginManager(manager interface{}) {
	// 플러그인 매니저에서 캐시 매니저 가져오기
	if mgr, ok := manager.(interface {
		GetCacheManager() *Manager
	}); ok {
		p.cacheManager = mgr.GetCacheManager()
	}
}

// GetCachedClient 캐시된 클라이언트 조회 또는 생성
func (p *CachePlugin) GetCachedClient(config map[string]interface{}) (client.Client, error) {
	if p.cacheManager == nil {
		return client.NewClient(), nil
	}

	// API 타입 감지
	apiType := "unknown"
	if kubeConfigStr, ok := config["kubeconfig"].(string); ok && kubeConfigStr != "" {
		apiType = "kubernetes"
	} else if minioEndpoint, ok := config["minio_endpoint"].(string); ok && minioEndpoint != "" {
		apiType = "minio"
	}

	// 새로운 캐시 매니저 사용
	return p.cacheManager.GetCachedClient(apiType, config)
}

// generateSimpleCacheKey 간단한 캐시 키 생성 (deprecated: manager.go의 메서드 사용)
func (p *CachePlugin) generateSimpleCacheKey(apiType string, config map[string]interface{}) string {
	if p.cacheManager != nil {
		return p.cacheManager.GetCacheKey(apiType, config)
	}

	// Fallback: 간단한 키 생성
	configStr := fmt.Sprintf("api:%s", apiType)
	if kubeConfig, ok := config["kubeconfig"].(string); ok && kubeConfig != "" {
		configStr += fmt.Sprintf("|kube:%s", kubeConfig)
	}
	if minioEndpoint, ok := config["minio_endpoint"].(string); ok && minioEndpoint != "" {
		configStr += fmt.Sprintf("|minio:%s", minioEndpoint)
	}

	hash := sha256.Sum256([]byte(configStr))
	return hex.EncodeToString(hash[:])
}

// GetCacheStats 캐시 통계 조회
func (p *CachePlugin) GetCacheStats() map[string]interface{} {
	if p.cacheManager == nil {
		return map[string]interface{}{
			"error": "cache not initialized",
		}
	}
	return p.cacheManager.GetStats()
}

// CleanupCache 캐시 정리
func (p *CachePlugin) CleanupCache() {
	if p.cacheManager != nil {
		p.cacheManager.Cleanup()
	}
}

// GetCacheKey 캐시 키 생성 (디버깅용)
func (p *CachePlugin) GetCacheKey(apiType string, config map[string]interface{}) string {
	return p.generateSimpleCacheKey(apiType, config)
}

// GetCacheInfo 특정 설정의 캐시 정보 조회
func (p *CachePlugin) GetCacheInfo(apiType string, config map[string]interface{}) map[string]interface{} {
	if p.cacheManager == nil {
		return map[string]interface{}{
			"error": "cache not initialized",
		}
	}

	cacheKey := p.GetCacheKey(apiType, config)

	return map[string]interface{}{
		"api_type":  apiType,
		"cache_key": cacheKey,
		"config":    config,
		"exists":    false, // 간단한 구현에서는 항상 false
		"status":    "cache_info_retrieved",
	}
}

// GetStatsHandler 캐시 통계 조회 핸들러
func (p *CachePlugin) GetStatsHandler(c echo.Context) error {
	stats := p.GetCacheStats()
	return response.RespondWithSuccessModel(c, 200, "Cache stats retrieved successfully", stats)
}

// CleanupHandler 캐시 정리 핸들러
func (p *CachePlugin) CleanupHandler(c echo.Context) error {
	p.CleanupCache()
	return response.RespondWithSuccessModel(c, 200, "Cache cleaned up successfully", nil)
}

// InvalidateHandler 캐시 무효화 핸들러
func (p *CachePlugin) InvalidateHandler(c echo.Context) error {
	var req struct {
		ApiType string                 `json:"api_type"`
		Config  map[string]interface{} `json:"config"`
	}

	if err := c.Bind(&req); err != nil {
		return response.RespondWithError(c, 400, "INVALID_REQUEST", "Invalid request body", err.Error())
	}

	// 새로운 캐시 매니저 사용
	if p.cacheManager != nil {
		p.cacheManager.Invalidate(req.ApiType, req.Config)
	}

	return response.RespondWithSuccessModel(c, 200, "Cache invalidated successfully", map[string]interface{}{
		"api_type": req.ApiType,
	})
}

// InvalidateAllHandler 모든 캐시 무효화 핸들러
func (p *CachePlugin) InvalidateAllHandler(c echo.Context) error {
	if p.cacheManager != nil {
		p.cacheManager.InvalidateAll()
	}

	return response.RespondWithSuccessModel(c, 200, "All cache invalidated successfully", nil)
}

// GetCacheInfoHandler 캐시 정보 조회 핸들러
func (p *CachePlugin) GetCacheInfoHandler(c echo.Context) error {
	var req struct {
		ApiType string                 `json:"api_type"`
		Config  map[string]interface{} `json:"config"`
	}

	if err := c.Bind(&req); err != nil {
		return response.RespondWithError(c, 400, "INVALID_REQUEST", "Invalid request body", err.Error())
	}

	// 캐시 정보 조회
	info := p.GetCacheInfo(req.ApiType, req.Config)

	return response.RespondWithSuccessModel(c, 200, "Cache info retrieved successfully", info)
}

// GetCacheKeyHandler 캐시 키 생성 핸들러 (디버깅용)
func (p *CachePlugin) GetCacheKeyHandler(c echo.Context) error {
	var req struct {
		ApiType string                 `json:"api_type"`
		Config  map[string]interface{} `json:"config"`
	}

	if err := c.Bind(&req); err != nil {
		return response.RespondWithError(c, 400, "INVALID_REQUEST", "Invalid request body", err.Error())
	}

	// 캐시 키 생성
	key := p.GetCacheKey(req.ApiType, req.Config)

	return response.RespondWithSuccessModel(c, 200, "Cache key generated successfully", map[string]interface{}{
		"api_type":  req.ApiType,
		"cache_key": key,
		"config":    req.Config,
	})
}
