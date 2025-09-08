package helm

import (
	"context"
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client/helm"
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

// GetCharts : Helm 차트 목록 조회
// @Summary Get Helm Charts
// @Description Get list of all Helm charts
// @Tags helm
// @Accept json
// @Produce json
// @Param namespace query string false "Namespace to get charts from" default(default)
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/helm/charts [get]
func (h *Handler) GetCharts(c echo.Context) error {
	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "default"
	}

	return h.HandleHelmResource(c, "helm-charts", func(helmClient helm.Client, ctx context.Context) (interface{}, error) {
		charts, err := helmClient.GetCharts(ctx, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get helm charts: %w", err)
		}

		return map[string]interface{}{
			"namespace": namespace,
			"charts":    charts,
			"count":     len(charts),
		}, nil
	})
}

// GetChart : 특정 Helm 차트 조회
// @Summary Get Helm Chart
// @Description Get specific Helm chart by name
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param namespace query string false "Namespace to get chart from" default(default)
// @Param version query string false "Chart version"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/helm/chart/{name} [get]
func (h *Handler) GetChart(c echo.Context) error {
	chartName := c.Param("name")
	namespace := c.QueryParam("namespace")
	versionStr := c.QueryParam("version")

	if namespace == "" {
		namespace = "default"
	}

	// version을 int로 변환 (기본값: 0)
	version := 0
	if versionStr != "" {
		if v, err := strconv.Atoi(versionStr); err == nil {
			version = v
		}
	}

	return h.HandleHelmResource(c, "helm-chart", func(helmClient helm.Client, ctx context.Context) (interface{}, error) {
		chart, err := helmClient.GetChart(ctx, chartName, namespace, version)
		if err != nil {
			return nil, fmt.Errorf("failed to get helm chart %s: %w", chartName, err)
		}

		return map[string]interface{}{
			"chart":     chart,
			"namespace": namespace,
			"version":   version,
		}, nil
	})
}

// IsChartInstalled : Helm 차트 설치 여부 확인
// @Summary Check Chart Installation Status
// @Description Check if Helm chart is installed
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/helm/chart/{name}/status [get]
func (h *Handler) IsChartInstalled(c echo.Context) error {
	chartName := c.Param("name")

	return h.HandleHelmResource(c, "helm-chart-status", func(helmClient helm.Client, ctx context.Context) (interface{}, error) {
		installed, release, err := helmClient.IsChartInstalled(chartName)
		if err != nil {
			return nil, fmt.Errorf("failed to check chart installation status for %s: %w", chartName, err)
		}

		result := map[string]interface{}{
			"chart_name": chartName,
			"installed":  installed,
		}

		if installed && release != nil {
			result["release"] = map[string]interface{}{
				"name":      release.Name,
				"namespace": release.Namespace,
				"version":   release.Chart.Metadata.Version,
				"status":    release.Info.Status,
			}
		}

		return result, nil
	})
}

// UninstallChart : Helm 차트 제거
// @Summary Uninstall Helm Chart
// @Description Uninstall Helm chart by name
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param namespace query string false "Namespace to uninstall chart from" default(default)
// @Param dryrun query boolean false "Dry run mode" default(false)
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/helm/chart/{name} [delete]
func (h *Handler) UninstallChart(c echo.Context) error {
	chartName := c.Param("name")
	namespace := c.QueryParam("namespace")
	dryrun := c.QueryParam("dryrun") == "true"

	if namespace == "" {
		namespace = "default"
	}

	return h.HandleHelmResource(c, "helm-chart-uninstall", func(helmClient helm.Client, ctx context.Context) (interface{}, error) {
		err := helmClient.UninstallChart(chartName, namespace, dryrun)
		if err != nil {
			return nil, fmt.Errorf("failed to uninstall helm chart %s: %w", chartName, err)
		}

		result := map[string]interface{}{
			"chart_name": chartName,
			"namespace":  namespace,
			"dryrun":     dryrun,
			"status":     "success",
		}

		if dryrun {
			result["message"] = fmt.Sprintf("Chart '%s' would be uninstalled from namespace '%s' (dry run)", chartName, namespace)
		} else {
			result["message"] = fmt.Sprintf("Chart '%s' has been uninstalled from namespace '%s'", chartName, namespace)
		}

		return result, nil
	})
}

// HealthCheck : Helm 연결 상태 확인
// @Summary Helm Health Check
// @Description Check Helm connection status
// @Tags helm
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/helm/health [get]
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.HandleHelmResource(c, "helm-health", func(helmClient helm.Client, ctx context.Context) (interface{}, error) {
		// Helm 연결 테스트 (헬스 체크)
		err := helmClient.HealthCheck(ctx)
		if err != nil {
			return nil, fmt.Errorf("helm health check failed: %w", err)
		}

		return map[string]interface{}{
			"service": "helm",
			"status":  "healthy",
			"message": "Helm connection is working",
		}, nil
	})
}
