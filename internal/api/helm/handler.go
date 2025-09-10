package helm

import (
	"context"
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/utils"
)

// Handler : Helm 관련 HTTP 핸들러
type Handler struct {
	*handler.BaseHandler
}

// NewHandler : 새로운 Helm 핸들러 생성
func NewHandler(base *handler.BaseHandler) *Handler {
	return &Handler{
		BaseHandler: base,
	}
}

// HealthCheck : Helm 연결 테스트
// @Summary Helm Connection Test
// @Description Test Helm connection with provided configuration
// @Tags helm
// @Accept json
// @Produce json
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/health [post]
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.HandleResourceClient(c, "helm-health", func(client client.Client, ctx context.Context) (interface{}, error) {

		// 네임스페이스 결정
		// "all"이면 모든 네임스페이스 조회
		namespace := utils.ResolveNamespace(c, "all")

		// Helm 연결 테스트
		_, err := client.Helm().GetCharts(ctx, namespace)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"service": "helm",
			"message": "Helm connection is working",
		}, nil
	})
}

// GetCharts : Helm 차트 목록 조회
// @Summary Get Helm Charts
// @Description Get list of all Helm charts
// @Tags helm
// @Accept json
// @Produce json
// @Param namespace query string false "Namespace name (default: 'default', all namespaces: 'all')"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/charts [get]
func (h *Handler) GetCharts(c echo.Context) error {
	return h.HandleResourceClient(c, "helm-charts", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		// "all"이면 모든 네임스페이스 조회,""이면 3번째 파라미터 값을 네임스페이스로 사용
		namespace := utils.ResolveNamespace(c, "all")

		// Helm 차트 목록 조회
		charts, err := client.Helm().GetCharts(ctx, namespace)
		if err != nil {
			return nil, err
		}

		return charts, nil
	})
}

// GetChart : 특정 Helm 차트 상세 조회
// @Summary Get Helm Chart Details
// @Description Get detailed information about a specific Helm chart
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param namespace query string false "Namespace name (default: 'default')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/charts/{name} [get]
func (h *Handler) GetChart(c echo.Context) error {
	return h.HandleResourceClient(c, "helm-chart", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "default")

		// 차트 이름 가져오기
		chartName := c.Param("name")
		if chartName == "" {
			return nil, echo.NewHTTPError(400, "chart name is required")
		}

		// 특정 차트 조회 (최신 버전)
		chart, err := client.Helm().GetChart(ctx, chartName, namespace, 0)
		if err != nil {
			return nil, err
		}

		return chart, nil
	})
}

// IsChartInstalled : 차트 설치 상태 확인
// @Summary Check Chart Installation Status
// @Description Check if a specific Helm chart is installed
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param namespace query string false "Namespace name (default: 'default')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/charts/{name}/status [get]
func (h *Handler) IsChartInstalled(c echo.Context) error {
	return h.HandleResourceClient(c, "helm-chart-status", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "default")

		// 차트 이름 가져오기
		chartName := c.Param("name")
		if chartName == "" {
			return nil, echo.NewHTTPError(400, "chart name is required")
		}

		// 차트 설치 상태 확인
		installed, release, err := client.Helm().IsChartInstalled(chartName)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"name":      chartName,
			"namespace": namespace,
			"installed": installed,
		}

		if installed && release != nil {
			result["release"] = release
		}

		return result, nil
	})
}

// InstallChart : Helm 차트 설치
// @Summary Install Helm Chart
// @Description Install Helm chart from URL (supports tgz URLs and versions)
// @Tags helm
// @Accept json
// @Produce json
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param releaseName query string true "Release name"
// @Param chartURL query string true "Chart URL (must be HTTP/HTTPS)"
// @Param version query string false "Chart version (optional)"
// @Param namespace query string false "Namespace name (default: 'default')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/charts [post]
func (h *Handler) InstallChart(c echo.Context) error {
	return h.HandleResourceClient(c, "helm-chart-install", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "default")

		// 파라미터 가져오기
		releaseName := c.QueryParam("releaseName")
		chartURL := c.QueryParam("chartURL")
		version := c.QueryParam("version")

		if releaseName == "" || chartURL == "" {
			return nil, echo.NewHTTPError(400, "releaseName and chartURL are required")
		}

		// Helm 차트 설치 (URL 기반, 버전 지원)
		err := client.Helm().InstallChart(releaseName, chartURL, version, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to install helm chart from %s: %w", chartURL, err)
		}

		result := map[string]interface{}{
			"release_name": releaseName,
			"chart_url":    chartURL,
			"namespace":    namespace,
			"status":       "installed",
		}

		if version != "" {
			result["version"] = version
		}

		return result, nil
	})
}

// UninstallChart : Helm 차트 제거
// @Summary Uninstall Helm Chart
// @Description Uninstall a specific Helm chart
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param namespace query string false "Namespace name (default: 'default')"
// @Param dryRun query boolean false "Dry run mode (default: false)"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/charts/{name} [delete]
func (h *Handler) UninstallChart(c echo.Context) error {
	return h.HandleResourceClient(c, "helm-chart-uninstall", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "default")

		// 차트 이름 가져오기
		chartName := c.Param("name")
		if chartName == "" {
			return nil, echo.NewHTTPError(400, "chart name is required")
		}

		// Dry run 모드 확인
		dryRunStr := c.QueryParam("dryRun")
		dryRun := false
		if dryRunStr != "" {
			var err error
			dryRun, err = strconv.ParseBool(dryRunStr)
			if err != nil {
				return nil, echo.NewHTTPError(400, "invalid dryRun parameter")
			}
		}

		// Helm 차트 제거
		err := client.Helm().UninstallChart(chartName, namespace, dryRun)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"name":      chartName,
			"namespace": namespace,
			"dry_run":   dryRun,
			"status":    "uninstalled",
		}

		return result, nil
	})
}

// UpgradeChart : Helm 차트 업그레이드
// @Summary Upgrade Helm Chart
// @Description Upgrade a specific Helm chart
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param chartPath query string true "Chart path or URL"
// @Param namespace query string false "Namespace name (default: 'default')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/charts/{name} [put]
func (h *Handler) UpgradeChart(c echo.Context) error {
	return h.HandleResourceClient(c, "helm-chart-upgrade", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "default")

		// 차트 이름 가져오기
		chartName := c.Param("name")
		if chartName == "" {
			return nil, echo.NewHTTPError(400, "chart name is required")
		}

		// 차트 경로 가져오기
		chartPath := c.QueryParam("chartPath")
		if chartPath == "" {
			return nil, echo.NewHTTPError(400, "chartPath is required")
		}

		// Helm 차트 업그레이드
		err := client.Helm().UpgradeChart(chartName, chartPath, nil)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"name":       chartName,
			"namespace":  namespace,
			"chart_path": chartPath,
			"status":     "upgraded",
		}

		return result, nil
	})
}

// GetChartHistory : 차트 히스토리 조회
// @Summary Get Chart History
// @Description Get installation history of a specific Helm chart
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param namespace query string false "Namespace name (default: 'default')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/charts/{name}/history [get]
func (h *Handler) GetChartHistory(c echo.Context) error {
	return h.HandleResourceClient(c, "helm-chart-history", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "default")

		// 차트 이름 가져오기
		chartName := c.Param("name")
		if chartName == "" {
			return nil, echo.NewHTTPError(400, "chart name is required")
		}

		// 차트 히스토리 조회 (최대 10개 버전)
		history := make([]interface{}, 0)
		for i := 1; i <= 10; i++ {
			chart, err := client.Helm().GetChart(ctx, chartName, namespace, i)
			if err != nil {
				// 더 이상 히스토리가 없으면 중단
				break
			}
			history = append(history, chart)
		}

		result := map[string]interface{}{
			"name":      chartName,
			"namespace": namespace,
			"history":   history,
		}

		return result, nil
	})
}

// GetChartValues : 차트 값 조회
// @Summary Get Chart Values
// @Description Get current values of a specific Helm chart
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param namespace query string false "Namespace name (default: 'default')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/charts/{name}/values [get]
func (h *Handler) GetChartValues(c echo.Context) error {
	return h.HandleResourceClient(c, "helm-chart-values", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "default")

		// 차트 이름 가져오기
		chartName := c.Param("name")
		if chartName == "" {
			return nil, echo.NewHTTPError(400, "chart name is required")
		}

		// 차트 조회 (최신 버전)
		chart, err := client.Helm().GetChart(ctx, chartName, namespace, 0)
		if err != nil {
			return nil, err
		}

		// 차트에서 값 추출
		result := map[string]interface{}{
			"name":      chartName,
			"namespace": namespace,
			"values":    chart,
		}

		return result, nil
	})
}
