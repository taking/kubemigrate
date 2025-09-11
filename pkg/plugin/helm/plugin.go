package helm

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/internal/validator"
	"github.com/taking/kubemigrate/pkg/client/helm"
	"github.com/taking/kubemigrate/pkg/errors"
	"github.com/taking/kubemigrate/pkg/utils"
	// "helm.sh/helm/v3/pkg/release" // 사용하지 않음
)

// HelmPlugin Helm 플러그인
type HelmPlugin struct {
	client    helm.Client
	config    map[string]interface{}
	validator *validator.KubernetesValidator
}

// NewPlugin 새로운 Helm 플러그인 생성
func NewPlugin() *HelmPlugin {
	return &HelmPlugin{
		validator: validator.NewKubernetesValidator(),
	}
}

// Name 플러그인 이름
func (p *HelmPlugin) Name() string {
	return "helm"
}

// Version 플러그인 버전
func (p *HelmPlugin) Version() string {
	return "1.0.0"
}

// Description 플러그인 설명
func (p *HelmPlugin) Description() string {
	return "Helm chart management plugin"
}

// Initialize 플러그인 초기화
func (p *HelmPlugin) Initialize(config map[string]interface{}) error {
	p.config = config

	// Helm 클라이언트 초기화 (기본 설정 사용)
	// TODO: 설정 기반 클라이언트 초기화 구현 필요
	p.client = helm.NewClient()

	return nil
}

// Shutdown 플러그인 종료
func (p *HelmPlugin) Shutdown() error {
	// 정리 작업이 필요한 경우 여기에 구현
	return nil
}

// RegisterRoutes 라우트 등록
func (p *HelmPlugin) RegisterRoutes(router *echo.Group) error {
	// Helm 관련 라우트 등록 (기존과 동일한 구조)
	helmGroup := router.Group("/helm")

	// 헬스체크
	helmGroup.POST("/health", p.HealthCheckHandler)

	// 차트 관리 (기존과 동일)
	helmGroup.POST("/charts", p.InstallChartHandler)                 // 차트 설치
	helmGroup.GET("/charts", p.GetChartsHandler)                     // 차트 목록 조회
	helmGroup.GET("/charts/:name", p.GetChartHandler)                // 차트 상세 조회
	helmGroup.GET("/charts/:name/status", p.IsChartInstalledHandler) // 차트 설치 상태
	helmGroup.PUT("/charts/:name", p.UpgradeChartHandler)            // 차트 업그레이드
	helmGroup.GET("/charts/:name/history", p.GetChartHistoryHandler) // 차트 히스토리 조회
	helmGroup.GET("/charts/:name/values", p.GetChartValuesHandler)   // 차트 값 조회
	helmGroup.DELETE("/charts/:name", p.UninstallChartHandler)       // 차트 제거

	return nil
}

// HealthCheck 헬스체크
func (p *HelmPlugin) HealthCheck(ctx context.Context) error {
	_, err := p.client.GetCharts(ctx, "default")
	return err
}

// GetServiceType 서비스 타입
func (p *HelmPlugin) GetServiceType() string {
	return "helm"
}

// GetClient 클라이언트 반환
func (p *HelmPlugin) GetClient() interface{} {
	return p.client
}

// HealthCheckHandler 헬스체크 핸들러
func (p *HelmPlugin) HealthCheckHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	// Helm 연결 테스트
	_, err := p.client.GetCharts(c.Request().Context(), "default")
	if err != nil {
		return errors.NewExternalError("helm", "GetCharts", err)
	}

	return response.RespondWithSuccessModel(c, 200, "Helm connection is working", map[string]interface{}{
		"service": "helm",
		"message": "Helm connection is working",
	})
}

// GetChartsHandler 차트 목록 조회 핸들러
func (p *HelmPlugin) GetChartsHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := utils.ResolveNamespace(c, "default")
	charts, err := p.client.GetCharts(c.Request().Context(), namespace)
	if err != nil {
		return errors.NewExternalError("helm", "GetCharts", err)
	}

	return response.RespondWithData(c, 200, charts)
}

// GetChartHandler 특정 차트 조회 핸들러
func (p *HelmPlugin) GetChartHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := utils.ResolveNamespace(c, "default")
	chartName := c.Param("name")

	// 기본 버전 0으로 조회
	chart, err := p.client.GetChart(c.Request().Context(), chartName, namespace, 0)
	if err != nil {
		return errors.NewExternalError("helm", "GetChart", err)
	}

	return response.RespondWithData(c, 200, chart)
}

// IsChartInstalledHandler 차트 설치 상태 확인 핸들러
func (p *HelmPlugin) IsChartInstalledHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	chartName := c.Param("name")

	installed, release, err := p.client.IsChartInstalled(chartName)
	if err != nil {
		return errors.NewExternalError("helm", "IsChartInstalled", err)
	}

	return response.RespondWithData(c, 200, map[string]interface{}{
		"chart":     chartName,
		"installed": installed,
		"release":   release,
	})
}

// GetChartHistoryHandler 차트 히스토리 조회 핸들러
func (p *HelmPlugin) GetChartHistoryHandler(c echo.Context) error {
	// 현재 클라이언트에서 지원하지 않는 기능
	return errors.NewValidationError("UNSUPPORTED_OPERATION", "Chart history not implemented", "GetChartHistory is not supported in current client")
}

// GetChartValuesHandler 차트 값 조회 핸들러
func (p *HelmPlugin) GetChartValuesHandler(c echo.Context) error {
	// 현재 클라이언트에서 지원하지 않는 기능
	return errors.NewValidationError("UNSUPPORTED_OPERATION", "Chart values not implemented", "GetChartValues is not supported in current client")
}

// InstallChartHandler 차트 설치 핸들러
func (p *HelmPlugin) InstallChartHandler(c echo.Context) error {
	var req struct {
		config.KubeConfig
		ChartName       string                 `json:"chartName"`
		ReleaseName     string                 `json:"releaseName"`
		Values          map[string]interface{} `json:"values,omitempty"`
		Version         string                 `json:"version,omitempty"`
		CreateNamespace bool                   `json:"createNamespace,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	chartName := c.Param("name")
	if chartName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing chart name", "chart name parameter is required")
	}

	// 기본값 설정
	if req.ReleaseName == "" {
		req.ReleaseName = chartName
	}

	err := p.client.InstallChart(req.ReleaseName, chartName, req.Version, req.Values)
	if err != nil {
		return errors.NewExternalError("helm", "InstallChart", err)
	}

	return response.RespondWithMessage(c, 201, "Chart installed successfully")
}

// UpgradeChartHandler 차트 업그레이드 핸들러
func (p *HelmPlugin) UpgradeChartHandler(c echo.Context) error {
	var req struct {
		config.KubeConfig
		ChartName   string                 `json:"chartName"`
		ReleaseName string                 `json:"releaseName"`
		Values      map[string]interface{} `json:"values,omitempty"`
		Version     string                 `json:"version,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	chartName := c.Param("name")
	if chartName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing chart name", "chart name parameter is required")
	}

	// 기본값 설정
	if req.ReleaseName == "" {
		req.ReleaseName = chartName
	}

	err := p.client.UpgradeChart(req.ReleaseName, chartName, req.Values)
	if err != nil {
		return errors.NewExternalError("helm", "UpgradeChart", err)
	}

	return response.RespondWithMessage(c, 200, "Chart upgraded successfully")
}

// UninstallChartHandler 차트 제거 핸들러
func (p *HelmPlugin) UninstallChartHandler(c echo.Context) error {
	var req struct {
		config.KubeConfig
		ReleaseName string `json:"releaseName"`
	}

	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	chartName := c.Param("name")
	if chartName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing chart name", "chart name parameter is required")
	}

	// 기본값 설정
	if req.ReleaseName == "" {
		req.ReleaseName = chartName
	}

	namespace := utils.ResolveNamespace(c, "default")

	err := p.client.UninstallChart(req.ReleaseName, namespace, false)
	if err != nil {
		return errors.NewExternalError("helm", "UninstallChart", err)
	}

	return response.RespondWithMessage(c, 200, "Chart uninstalled successfully")
}
