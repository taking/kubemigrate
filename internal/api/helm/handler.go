package helm

import (
	"context"
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client/helm"
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
// @Param request body config.HelmConfig true "Helm configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/health [post]
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.HandleHelmResource(c, "helm-health", func(helmClient helm.Client, ctx context.Context) (interface{}, error) {
		// Helm 연결 테스트
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
	return h.HandleHelmResource(c, "helm-charts", func(helmClient helm.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		_, err := utils.BindAndValidateHelmConfig(c, h.KubernetesValidator)
		if err != nil {
			return nil, err
		}

		// 네임스페이스 결정
		// "all"이면 모든 네임스페이스 조회,""이면 3번째 파라미터 값을 네임스페이스로 사용
		namespace := utils.ResolveNamespace(c, "default")

		// Helm 차트 목록 조회
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
// @Param namespace query string false "Namespace name (default: 'default', all namespaces: 'all')"
// @Param version query string false "Chart Release version"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/chart/{name} [get]
func (h *Handler) GetChart(c echo.Context) error {
	return h.HandleHelmResource(c, "helm-chart", func(helmClient helm.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		_, err := utils.BindAndValidateHelmConfig(c, h.KubernetesValidator)
		if err != nil {
			return nil, err
		}

		// 네임스페이스 결정
		// "all"이면 모든 네임스페이스 조회,""이면 3번째 파라미터 값을 네임스페이스로 사용
		namespace := utils.ResolveNamespace(c, "default")

		chartName := c.Param("name")
		versionStr := c.QueryParam("version")

		// version을 int로 변환 (기본값: 0)
		version := 0
		if versionStr != "" {
			if v, err := strconv.Atoi(versionStr); err == nil {
				version = v
			}
		}

		// Helm 차트 단일 조회
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
// @Router /v1/helm/chart/{name}/status [get]
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
// @Param namespace query string false "Namespace name (default: 'default', all namespaces: 'all')"
// @Param dryrun query boolean false "Dry run mode" default(false)
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/helm/chart/{name} [delete]
func (h *Handler) UninstallChart(c echo.Context) error {
	return h.HandleHelmResource(c, "helm-chart-uninstall", func(helmClient helm.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		_, err := utils.BindAndValidateHelmConfig(c, h.KubernetesValidator)
		if err != nil {
			return nil, err
		}

		// 네임스페이스 결정
		// "all"이면 모든 네임스페이스 조회,""이면 3번째 파라미터 값을 네임스페이스로 사용
		namespace := utils.ResolveNamespace(c, "default")

		chartName := c.Param("name")
		dryrun := c.QueryParam("dryrun") == "true"

		// Helm 차트 제거
		err = helmClient.UninstallChart(chartName, namespace, dryrun)
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
