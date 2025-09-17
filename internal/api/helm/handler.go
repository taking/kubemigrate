package helm

import (
	"context"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
	"github.com/taking/kubemigrate/pkg/utils"
)

// Handler : Helm 관련 HTTP 핸들러
type Handler struct {
	*handler.BaseHandler
	service *Service
}

// NewHandler : 새로운 Helm 핸들러 생성
func NewHandler(base *handler.BaseHandler) *Handler {
	return &Handler{
		BaseHandler: base,
		service:     NewService(),
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
		namespace := utils.ResolveNamespace(c, "all")

		// 서비스 로직 호출
		return h.service.HealthCheckInternal(client, ctx, namespace)
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
		namespace := utils.ResolveNamespace(c, "all")

		// 서비스 로직 호출
		return h.service.GetChartsInternal(client, ctx, namespace)
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

		// 서비스 로직 호출
		return h.service.GetChartInternal(client, ctx, chartName, namespace, 0)
	})
}

// GetChartStatus : 차트 상태 조회
// @Summary Get Helm Chart Status
// @Description Get status information about a specific Helm chart
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param namespace query string false "Namespace name (default: 'default')"
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/charts/{name}/status [get]
func (h *Handler) GetChartStatus(c echo.Context) error {
	return h.HandleResourceClient(c, "helm-chart-status", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "default")

		// 차트 이름 가져오기
		chartName := c.Param("name")
		if chartName == "" {
			return nil, echo.NewHTTPError(400, "chart name is required")
		}

		// 서비스 로직 호출
		return h.service.GetChartStatusInternal(client, ctx, chartName, namespace)
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

		// 서비스 로직 호출
		config := config.InstallChartConfig{
			ReleaseName: releaseName,
			ChartURL:    chartURL,
			Version:     version,
			Namespace:   namespace,
			Values:      nil,
		}

		return h.service.InstallChartInternal(client, ctx, config)
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

		// 서비스 로직 호출
		return h.service.UninstallChartInternal(client, ctx, chartName, namespace, dryRun)
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

		// 서비스 로직 호출
		config := config.UpgradeChartConfig{
			ReleaseName: chartName,
			ChartPath:   chartPath,
			Namespace:   namespace,
			Values:      nil,
		}

		return h.service.UpgradeChartInternal(client, ctx, config)
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

		// 서비스 로직 호출
		return h.service.GetChartHistoryInternal(client, ctx, chartName, namespace)
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

		// 서비스 로직 호출
		return h.service.GetChartValuesInternal(client, ctx, chartName, namespace)
	})
}
