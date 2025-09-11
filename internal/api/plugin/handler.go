package plugin

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/pkg/errors"
	pluginpkg "github.com/taking/kubemigrate/pkg/plugin"
)

// Handler 플러그인 관리 핸들러
type Handler struct {
	*handler.BaseHandler
	pluginManager *pluginpkg.Manager
}

// NewHandler 새로운 플러그인 핸들러 생성
func NewHandler(base *handler.BaseHandler, pluginManager *pluginpkg.Manager) *Handler {
	return &Handler{
		BaseHandler:   base,
		pluginManager: pluginManager,
	}
}

// ListPlugins 등록된 플러그인 목록 조회
// @Summary List Plugins
// @Description Get list of all registered plugins
// @Tags plugin
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /v1/plugins [get]
func (h *Handler) ListPlugins(c echo.Context) error {
	plugins := h.pluginManager.ListPlugins()
	return response.RespondWithData(c, http.StatusOK, plugins)
}

// GetPlugin 특정 플러그인 정보 조회
// @Summary Get Plugin Info
// @Description Get information about a specific plugin
// @Tags plugin
// @Accept json
// @Produce json
// @Param name path string true "Plugin name"
// @Success 200 {object} response.SuccessResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /v1/plugins/{name} [get]
func (h *Handler) GetPlugin(c echo.Context) error {
	pluginName := c.Param("name")
	if pluginName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing plugin name", "plugin name parameter is required")
	}

	plugin, exists := h.pluginManager.GetPlugin(pluginName)
	if !exists {
		return errors.NewValidationError(errors.CodeResourceNotFound, "Plugin not found", "plugin not found")
	}

	pluginInfo := pluginpkg.PluginInfo{
		Name:        plugin.Name(),
		Version:     plugin.Version(),
		Description: plugin.Description(),
		Enabled:     true, // 기본값
	}

	// 서비스 타입 추가
	if servicePlugin, ok := plugin.(pluginpkg.ServicePlugin); ok {
		pluginInfo.ServiceType = servicePlugin.GetServiceType()
	}

	return response.RespondWithData(c, http.StatusOK, pluginInfo)
}

// HealthCheckAllPlugins 모든 플러그인 헬스체크
// @Summary Health Check All Plugins
// @Description Check health status of all plugins
// @Tags plugin
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /v1/plugins/health [get]
func (h *Handler) HealthCheckAllPlugins(c echo.Context) error {
	results := h.pluginManager.HealthCheckAllPlugins(c.Request().Context())

	// Spring Boot Actuator 스타일로 결과 정리
	components := make(map[string]interface{})
	for name, err := range results {
		if err != nil {
			components[name] = map[string]interface{}{
				"status": "DOWN",
				"details": map[string]interface{}{
					"error": err.Error(),
				},
			}
		} else {
			components[name] = map[string]interface{}{
				"status": "UP",
			}
		}
	}

	// 전체 상태 결정 (모든 컴포넌트가 UP이면 UP, 하나라도 DOWN이면 DOWN)
	overallStatus := "UP"
	for _, component := range components {
		if comp, ok := component.(map[string]interface{}); ok {
			if status, exists := comp["status"]; exists && status == "DOWN" {
				overallStatus = "DOWN"
				break
			}
		}
	}

	healthData := map[string]interface{}{
		"status":     overallStatus,
		"components": components,
	}

	return response.RespondWithData(c, http.StatusOK, healthData)
}

// HealthCheckPlugin 특정 플러그인 헬스체크
// @Summary Health Check Plugin
// @Description Check health status of a specific plugin
// @Tags plugin
// @Accept json
// @Produce json
// @Param name path string true "Plugin name"
// @Success 200 {object} response.SuccessResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /v1/plugins/{name}/health [get]
func (h *Handler) HealthCheckPlugin(c echo.Context) error {
	pluginName := c.Param("name")
	if pluginName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing plugin name", "plugin name parameter is required")
	}

	plugin, exists := h.pluginManager.GetPlugin(pluginName)
	if !exists {
		return errors.NewValidationError(errors.CodeResourceNotFound, "Plugin not found", "plugin not found")
	}

	err := plugin.HealthCheck(c.Request().Context())
	if err != nil {
		healthData := map[string]interface{}{
			"status": "DOWN",
			"components": map[string]interface{}{
				pluginName: map[string]interface{}{
					"status": "DOWN",
					"details": map[string]interface{}{
						"error": err.Error(),
					},
				},
			},
		}
		return response.RespondWithData(c, http.StatusOK, healthData)
	}

	healthData := map[string]interface{}{
		"status": "UP",
		"components": map[string]interface{}{
			pluginName: map[string]interface{}{
				"status": "UP",
			},
		},
	}
	return response.RespondWithData(c, http.StatusOK, healthData)
}
