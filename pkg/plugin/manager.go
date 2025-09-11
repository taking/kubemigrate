package plugin

import (
	"context"
	"fmt"
	"sync"

	"github.com/labstack/echo/v4"
)

// Manager 플러그인 매니저
type Manager struct {
	plugins map[string]Plugin
	configs map[string]PluginConfig
	mutex   sync.RWMutex
}

// NewManager 새로운 플러그인 매니저 생성
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
		configs: make(map[string]PluginConfig),
	}
}

// RegisterPlugin 플러그인 등록
func (m *Manager) RegisterPlugin(plugin Plugin) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.plugins[plugin.Name()]; exists {
		return fmt.Errorf("plugin %s already registered", plugin.Name())
	}

	m.plugins[plugin.Name()] = plugin
	return nil
}

// SetPluginConfig 플러그인 설정
func (m *Manager) SetPluginConfig(name string, config PluginConfig) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.configs[name] = config
}

// InitializePlugin 플러그인 초기화
func (m *Manager) InitializePlugin(name string) error {
	m.mutex.RLock()
	plugin, exists := m.plugins[name]
	config, configExists := m.configs[name]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// 설정이 없으면 기본 설정 사용
	if !configExists {
		config = PluginConfig{Enabled: true, Config: make(map[string]interface{})}
	}

	// 플러그인이 비활성화되어 있으면 초기화하지 않음
	if !config.Enabled {
		return nil
	}

	return plugin.Initialize(config.Config)
}

// InitializeAllPlugins 모든 플러그인 초기화
func (m *Manager) InitializeAllPlugins() error {
	m.mutex.RLock()
	pluginNames := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		pluginNames = append(pluginNames, name)
	}
	m.mutex.RUnlock()

	for _, name := range pluginNames {
		if err := m.InitializePlugin(name); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
		}
	}

	return nil
}

// RegisterAllRoutes 모든 플러그인의 라우트 등록
func (m *Manager) RegisterAllRoutes(router *echo.Group) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for name, plugin := range m.plugins {
		// 플러그인이 비활성화되어 있으면 라우트 등록하지 않음
		if config, exists := m.configs[name]; exists && !config.Enabled {
			continue
		}

		if err := plugin.RegisterRoutes(router); err != nil {
			return fmt.Errorf("failed to register routes for plugin %s: %w", name, err)
		}
	}

	return nil
}

// GetPlugin 플러그인 조회
func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	plugin, exists := m.plugins[name]
	return plugin, exists
}

// GetServicePlugin 서비스 플러그인 조회
func (m *Manager) GetServicePlugin(serviceType string) (ServicePlugin, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, plugin := range m.plugins {
		if servicePlugin, ok := plugin.(ServicePlugin); ok {
			if servicePlugin.GetServiceType() == serviceType {
				return servicePlugin, true
			}
		}
	}

	return nil, false
}

// ListPlugins 등록된 플러그인 목록 조회
func (m *Manager) ListPlugins() []PluginInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plugins := make([]PluginInfo, 0, len(m.plugins))
	for name, plugin := range m.plugins {
		info := PluginInfo{
			Name:        plugin.Name(),
			Version:     plugin.Version(),
			Description: plugin.Description(),
			Enabled:     true, // 기본값
		}

		// 설정에서 활성화 상태 확인
		if config, exists := m.configs[name]; exists {
			info.Enabled = config.Enabled
		}

		// 서비스 타입 추가
		if servicePlugin, ok := plugin.(ServicePlugin); ok {
			info.ServiceType = servicePlugin.GetServiceType()
		}

		plugins = append(plugins, info)
	}

	return plugins
}

// ShutdownAllPlugins 모든 플러그인 종료
func (m *Manager) ShutdownAllPlugins() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var lastErr error
	for name, plugin := range m.plugins {
		if err := plugin.Shutdown(); err != nil {
			lastErr = fmt.Errorf("failed to shutdown plugin %s: %w", name, err)
		}
	}

	return lastErr
}

// HealthCheckAllPlugins 모든 플러그인 헬스체크
func (m *Manager) HealthCheckAllPlugins(ctx context.Context) map[string]error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	results := make(map[string]error)
	for name, plugin := range m.plugins {
		// 플러그인이 비활성화되어 있으면 헬스체크하지 않음
		if config, exists := m.configs[name]; exists && !config.Enabled {
			continue
		}

		results[name] = plugin.HealthCheck(ctx)
	}

	return results
}
