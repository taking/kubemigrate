package plugin

import (
	"context"

	"github.com/labstack/echo/v4"
)

// Plugin 인터페이스 - 모든 플러그인이 구현해야 하는 기본 인터페이스
type Plugin interface {
	// 플러그인 메타데이터
	Name() string
	Version() string
	Description() string

	// 플러그인 생명주기
	Initialize(config map[string]interface{}) error
	Shutdown() error

	// HTTP 라우트 등록
	RegisterRoutes(router *echo.Group) error

	// 헬스체크
	HealthCheck(ctx context.Context) error

	// 플러그인 매니저 설정 (캐시 사용을 위해)
	SetPluginManager(manager interface{})
}

// ServicePlugin - 서비스별 플러그인 인터페이스
type ServicePlugin interface {
	Plugin

	// 서비스별 특화 메서드
	GetServiceType() string
	GetClient() interface{}
}

// PluginConfig 플러그인 설정 구조체
type PluginConfig struct {
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

// PluginInfo 플러그인 정보 구조체
type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	ServiceType string `json:"service_type,omitempty"`
}
