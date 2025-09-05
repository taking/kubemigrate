package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/pkg/cache"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/health"
	"github.com/taking/kubemigrate/pkg/interfaces"
	_ "github.com/taking/kubemigrate/pkg/models"
	"github.com/taking/kubemigrate/pkg/response"
	"github.com/taking/kubemigrate/pkg/utils"
	"github.com/taking/kubemigrate/pkg/validator"
)

// HelmHandler : Helm 관련 HTTP 요청을 처리하는 핸들러
type HelmHandler struct {
	kubernetesValidator *validator.KubernetesValidator
	cache               *cache.Cache
	workerPool          *utils.WorkerPool
	healthManager       *health.HealthManager
}

// NewHelmHandler : 새로운 HelmHandler 인스턴스 생성
func NewHelmHandler(appCache *cache.Cache, workerPool *utils.WorkerPool, healthManager *health.HealthManager) *HelmHandler {
	return &HelmHandler{
		kubernetesValidator: validator.NewKubernetesValidator(),
		cache:               appCache,
		workerPool:          workerPool,
		healthManager:       healthManager,
	}
}

// HealthCheck : Helm 연결 상태 확인
// @Summary Helm 연결 상태 확인
// @Description Helm 연결 상태를 확인합니다
// @Tags helm
// @Accept json
// @Produce json
// @Param request body models.HelmConfigRequest true "Helm connection configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Connection successful"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Failure 503 {object} models.SwaggerErrorResponse "Service unavailable"
// @Router /helm/health [get]
func (h *HelmHandler) HealthCheck(c echo.Context) error {
	req, err := utils.BindAndValidateKubeConfig(c, h.kubernetesValidator)
	if err != nil {
		return err
	}

	// 기본 네임스페이스 설정
	req.Namespace = utils.ResolveNamespace(&req, c, "default")

	client, err := client.NewHelmClient(req)
	if err != nil {
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}

	// 클러스터 연결 상태 확인
	if err := client.HealthCheck(c.Request().Context()); err != nil {
		return response.RespondError(c, http.StatusServiceUnavailable,
			"Helm cluster unhealthy: "+err.Error())
	}

	// 헬스체크 매니저에 Helm 체커 등록
	if h.healthManager != nil {
		helmChecker := health.NewHelmHealthChecker(req)
		h.healthManager.Register(helmChecker)
	}

	return response.RespondStatus(c, "healthy", "Helm connection successful")
}

// GetCharts : Helm 차트 목록 조회
// @Summary Helm 차트 목록 조회
// @Description Helm 차트 목록을 조회합니다
// @Tags helm
// @Accept json
// @Produce json
// @Success 200 {object} models.SwaggerSuccessResponse "Charts retrieved"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /helm/charts [get]
func (h *HelmHandler) GetCharts(c echo.Context) error {
	return h.handleHelmResourceWithCache(c, "charts", func(client interfaces.HelmClient, ctx context.Context) (interface{}, error) {
		charts, err := client.GetCharts(ctx)
		if err != nil {
			return nil, err
		}
		return charts, nil
	},
	)
}

// GetChart : Helm 차트 조회
// @Summary Helm 차트 조회
// @Description Helm 차트를 조회합니다
// @Tags helm
// @Accept json
// @Produce json
// @Param name path string true "Chart name"
// @Param namespace query string false "Namespace"
// @Param version query int false "Release Version"
// @Success 200 {object} models.SwaggerSuccessResponse "Chart retrieved"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /helm/chart/:name [get]
func (h *HelmHandler) GetChart(c echo.Context) error {
	releaseName := c.Param("name")

	return h.handleHelmResourceWithCache(c, fmt.Sprintf("chart-%s", releaseName), func(client interfaces.HelmClient, ctx context.Context) (interface{}, error) {
		namespace := c.QueryParam("namespace") // optional
		version := c.QueryParam("version")     // optional

		releaseVersion := utils.StringToIntOrDefault(version, 1) // 기본값 (releaseVersion 1)

		chart, err := client.GetChart(ctx, releaseName, namespace, releaseVersion)
		if err != nil {
			return nil, err
		}
		return chart, nil
	},
	)
}

// IsChartInstalled : Helm 차트 설치 여부 확인
// @Summary Helm 차트 설치 여부 확인
// @Description Helm 차트 설치 여부를 확인합니다
// @Tags helm
// @Accept json
// @Produce json
// @Param request body models.HelmConfigRequest true "Helm chart check configuration"
// @Param name query string false "Release Name"
// @Success 200 {object} models.SwaggerSuccessResponse "Chart status retrieved"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /helm/chart/:name/status [get]
func (h *HelmHandler) IsChartInstalled(c echo.Context) error {
	releaseName := c.Param("name")

	return h.handleHelmResourceWithCache(c, fmt.Sprintf("chart-installed-%s", releaseName), func(client interfaces.HelmClient, ctx context.Context) (interface{}, error) {
		installed, _, err := client.IsChartInstalled(releaseName)
		if err != nil {
			return nil, err
		}

		status := "not_installed"
		message := "Chart is not installed"
		if installed {
			status = "installed"
			message = "Chart is installed"
		}

		return map[string]interface{}{
			"status":  status,
			"message": message,
		}, nil
	},
	)
}

// InstallChart : Helm 차트 설치
// @Summary Helm 차트 설치
// @Description Helm 차트를 설치합니다
// @Tags helm
// @Accept json
// @Produce json
// @Param request body models.HelmConfigRequest true "Helm chart installation configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Chart installed"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /helm/chart [post]
// func (h *HelmHandler) InstallChart(c echo.Context) error {
// 	releaseName := c.Param("name")

// 	return h.handleHelmResourceWithCache(c, "chart", func(client interfaces.HelmClient, ctx context.Context) (interface{}, error) {
// 		err := client.InstallChart(releaseName, chartPath, values)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return nil, nil
// 	},
// 	)
// }

// UninstallChart : Helm 차트 삭제
// @Summary Helm 차트 삭제
// @Description Helm 차트를 삭제합니다
// @Tags helm
// @Accept json
// @Produce json
// @Param request body models.HelmConfigRequest true "Helm chart uninstall configuration"
// @Param name query string false "Release Name"
// @Param namespace query string false "Namespace"
// @Param dryrun query bool false "Dry Run"
// @Success 200 {object} models.SwaggerSuccessResponse "Chart uninstalled"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /helm/chart/:name [delete]
func (h *HelmHandler) UninstallChart(c echo.Context) error {
	return h.handleHelmResourceWithCache(c, "chart", func(client interfaces.HelmClient, ctx context.Context) (interface{}, error) {
		releaseName := c.Param("name")
		namespace := c.QueryParam("namespace")
		dryRun := utils.StringToBoolOrDefault(c.QueryParam("dryrun"), true)

		err := client.UninstallChart(releaseName, namespace, dryRun)
		if err != nil && !dryRun {
			return nil, err
		}

		return map[string]interface{}{
			"message": err,
			"dryrun":  dryRun,
		}, nil
	})
}

// handleHelmResourceWithCache : 캐시를 사용하는 Helm 리소스 처리 헬퍼
func (h *HelmHandler) handleHelmResourceWithCache(c echo.Context, cacheKey string,
	getResource func(interfaces.HelmClient, context.Context) (interface{}, error)) error {

	// HelmConfig 검증
	req, err := utils.BindAndValidateHelmConfig(c, h.kubernetesValidator)
	if err != nil {
		return err
	}

	// 기본 네임스페이스 설정
	req.Namespace = utils.ResolveNamespace(&req.KubeConfig, c, "default")

	// 캐시 키 생성
	fullCacheKey := fmt.Sprintf("helm:%s:%s", cacheKey, req.Namespace)

	// 캐시에서 가져오기
	if cached, exists := h.cache.Get(fullCacheKey); exists {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   cached,
			"cached": true,
		})
	}

	// 캐시에 없으면 클라이언트에서 가져오기
	helmClient, err := client.NewHelmClient(req.KubeConfig)
	if err != nil {
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}

	// 워커 풀을 사용하여 백그라운드에서 데이터 가져오기
	resultChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)

	h.workerPool.Submit(func() {
		data, err := getResource(helmClient, c.Request().Context())
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- data
	})

	select {
	case data := <-resultChan:
		// 캐시에 저장
		h.cache.Set(fullCacheKey, data)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   data,
			"cached": false,
		})
	case err := <-errorChan:
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}
}
