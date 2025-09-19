package helm

import (
	"context"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/response"
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
		service:     NewService(base),
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
	return h.BaseHandler.HealthCheck(c, handler.HealthCheckConfig{
		ServiceName: "helm",
		DefaultNS:   "all",
		HealthFunc: func(client client.Client, ctx context.Context) error {
			namespace := h.ResolveNamespace(c, "all")
			_, err := client.Helm().GetCharts(ctx, namespace)
			return err
		},
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
		namespace := h.ResolveNamespace(c, "all")

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
		namespace := h.ResolveNamespace(c, "default")

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
		namespace := h.ResolveNamespace(c, "default")

		// 차트 이름 가져오기
		chartName := c.Param("name")
		if chartName == "" {
			return nil, echo.NewHTTPError(400, "chart name is required")
		}

		// 서비스 로직 호출
		return h.service.GetChartStatusInternal(client, ctx, chartName, namespace)
	})
}

// InstallChart : Helm 차트 비동기 설치
// @Summary Install Helm Chart Asynchronously
// @Description Install a Helm chart asynchronously and return job ID for status tracking
// @Tags helm
// @Accept json
// @Produce json
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param releaseName query string true "Release name"
// @Param chartURL query string true "Chart URL (must be HTTP/HTTPS)"
// @Param version query string false "Chart version (optional)"
// @Param namespace query string false "Namespace name (default: 'default')"
// @Param values query string false "Chart values as JSON string"
// @Success 200 {object} map[string]interface{} "Job started"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/helm/charts [post]
func (h *Handler) InstallChart(c echo.Context) error {
	// Query parameters 파싱
	releaseName := c.QueryParam("releaseName")
	chartURL := c.QueryParam("chartURL")
	version := c.QueryParam("version")
	namespace := h.ResolveNamespace(c, "default")
	valuesStr := c.QueryParam("values")

	// 필수 파라미터 검증
	if releaseName == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "releaseName is required", "")
	}
	if chartURL == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "chartURL is required", "")
	}
	if version == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "version is required", "")
	}

	// Values 파싱
	var values map[string]interface{}
	if valuesStr != "" {
		if err := utils.ParseJSON(valuesStr, &values); err != nil {
			return response.RespondWithErrorModel(c, 400, "INVALID_JSON", "Invalid values JSON format", err.Error())
		}
	}

	// Config 생성
	config := config.InstallChartConfig{
		ReleaseName: releaseName,
		ChartURL:    chartURL,
		Version:     version,
		Namespace:   namespace,
		Values:      values,
	}

	return h.HandleResourceClient(c, "helm-install-async", func(client client.Client, ctx context.Context) (interface{}, error) {
		return h.service.InstallChartAsyncInternal(client, ctx, config)
	})
}

// UninstallChart : Helm 차트 비동기 제거
// @Summary Uninstall Helm Chart Asynchronously
// @Description Uninstall a specific Helm chart asynchronously and return job ID for status tracking
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param namespace query string false "Namespace name (default: 'default')"
// @Param dryRun query boolean false "Dry run mode (default: false)"
// @Success 200 {object} map[string]interface{} "Job started"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/helm/charts/{name} [delete]
func (h *Handler) UninstallChart(c echo.Context) error {
	// 네임스페이스 결정
	namespace := h.ResolveNamespace(c, "default")

	// 차트 이름 가져오기
	chartName := c.Param("name")
	if chartName == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "chart name is required", "")
	}

	// Dry run 모드 확인
	dryRunStr := c.QueryParam("dryRun")
	dryRun := false
	if dryRunStr != "" {
		var err error
		dryRun, err = strconv.ParseBool(dryRunStr)
		if err != nil {
			return response.RespondWithErrorModel(c, 400, "INVALID_PARAMETER", "invalid dryRun parameter", err.Error())
		}
	}

	return h.HandleResourceClient(c, "helm-uninstall-async", func(client client.Client, ctx context.Context) (interface{}, error) {
		return h.service.UninstallChartAsyncInternal(client, ctx, chartName, namespace, dryRun)
	})
}

// UpgradeChart : Helm 차트 비동기 업그레이드
// @Summary Upgrade Helm Chart Asynchronously
// @Description Upgrade a specific Helm chart asynchronously and return job ID for status tracking
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param chartURL query string false "Chart URL (HTTP/HTTPS)"
// @Param chartPath query string false "Chart local path"
// @Param version query string false "Chart version"
// @Param namespace query string false "Namespace name (default: 'default')"
// @Param values query string false "Chart values as JSON string"
// @Success 200 {object} map[string]interface{} "Job started"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/helm/charts/{name} [put]
func (h *Handler) UpgradeChart(c echo.Context) error {
	// 네임스페이스 결정
	namespace := h.ResolveNamespace(c, "default")

	// 차트 이름 가져오기
	chartName := c.Param("name")
	if chartName == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "chart name is required", "")
	}

	// 차트 URL/경로 가져오기 (InstallChart와 동일한 방식)
	chartURL := c.QueryParam("chartURL")
	chartPath := c.QueryParam("chartPath")

	// chartURL 또는 chartPath 중 하나는 필수
	if chartURL == "" && chartPath == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "chartURL or chartPath is required", "")
	}

	// chartURL이 있으면 URL 방식, 없으면 로컬 경로 방식
	var finalChartPath string
	if chartURL != "" {
		finalChartPath = chartURL
	} else {
		finalChartPath = chartPath
	}

	// Values 파싱
	valuesStr := c.QueryParam("values")
	var values map[string]interface{}
	if valuesStr != "" {
		if err := utils.ParseJSON(valuesStr, &values); err != nil {
			return response.RespondWithErrorModel(c, 400, "INVALID_JSON", "Invalid values JSON format", err.Error())
		}
	}

	// Config 생성
	config := config.UpgradeChartConfig{
		ReleaseName: chartName,
		ChartPath:   finalChartPath,
		Namespace:   namespace,
		Values:      values,
	}

	return h.HandleResourceClient(c, "helm-upgrade-async", func(client client.Client, ctx context.Context) (interface{}, error) {
		return h.service.UpgradeChartAsyncInternal(client, ctx, config)
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
		namespace := h.ResolveNamespace(c, "default")

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
		namespace := h.ResolveNamespace(c, "default")

		// 차트 이름 가져오기
		chartName := c.Param("name")
		if chartName == "" {
			return nil, echo.NewHTTPError(400, "chart name is required")
		}

		// 서비스 로직 호출
		return h.service.GetChartValuesInternal(client, ctx, chartName, namespace)
	})
}

// GetJobStatus : 작업 상태 조회
// @Summary Get Job Status
// @Description Get the status of a specific job
// @Tags helm
// @Accept json
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} map[string]interface{} "Job status"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/helm/charts/status/{jobId} [get]
func (h *Handler) GetJobStatus(c echo.Context) error {
	jobID := c.Param("jobId")
	if jobID == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "jobId is required", "")
	}

	result, err := h.service.GetJobStatusInternal(jobID)
	if err != nil {
		return response.RespondWithErrorModel(c, 404, "JOB_NOT_FOUND", err.Error(), "")
	}

	return response.RespondWithData(c, 200, result)
}

// GetJobLogs : 작업 로그 조회
// @Summary Get Job Logs
// @Description Get the logs of a specific job
// @Tags helm
// @Accept json
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} map[string]interface{} "Job logs"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/helm/charts/logs/{jobId} [get]
func (h *Handler) GetJobLogs(c echo.Context) error {
	jobID := c.Param("jobId")
	if jobID == "" {
		return response.RespondWithErrorModel(c, 400, "MISSING_PARAMETER", "jobId is required", "")
	}

	result, err := h.service.GetJobLogsInternal(jobID)
	if err != nil {
		return response.RespondWithErrorModel(c, 404, "JOB_NOT_FOUND", err.Error(), "")
	}

	return response.RespondWithData(c, 200, result)
}

// GetAllJobs : 모든 작업 조회
// @Summary Get All Jobs
// @Description Get all jobs
// @Tags helm
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "All jobs"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/helm/charts/jobs [get]
func (h *Handler) GetAllJobs(c echo.Context) error {
	result, err := h.service.GetAllJobsInternal()
	if err != nil {
		return response.RespondWithErrorModel(c, 500, "INTERNAL_ERROR", err.Error(), "")
	}

	return response.RespondWithData(c, 200, result)
}
