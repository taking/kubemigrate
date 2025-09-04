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
	"github.com/taking/kubemigrate/pkg/models"
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
// @Summary Check Helm connection
// @Description Validate Helm connection using KubeConfig
// @Tags helm
// @Accept json
// @Produce json
// @Param request body models.KubeConfigRequest true "Helm connection configuration"
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

// IsChartInstalled : Helm 차트 설치 여부 확인
// @Summary Check if Helm chart is installed
// @Description Check if specified Helm chart is installed using KubeConfig
// @Tags helm
// @Accept json
// @Produce json
// @Param request body models.HelmChartRequest true "Helm chart check configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Chart status retrieved"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /helm/chart_check [post]
func (h *HelmHandler) IsChartInstalled(c echo.Context) error {
	var req models.HelmChartRequest
	if err := c.Bind(&req); err != nil {
		return response.RespondError(c, http.StatusBadRequest, "invalid request body")
	}

	// KubeConfig 검증
	decodedKubeConfig, err := h.kubernetesValidator.ValidateKubernetesConfig(&req.KubeConfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	req.KubeConfig.KubeConfig = decodedKubeConfig

	// 기본 네임스페이스 설정
	req.Namespace = utils.ResolveNamespace(&req.KubeConfig, c, "default")

	return h.handleHelmResourceWithCache(c, fmt.Sprintf("chart-status-%s", req.ChartName), req.KubeConfig,
		func(client interfaces.HelmClient, ctx context.Context) (interface{}, error) {
			installed, _, err := client.IsChartInstalled(req.ChartName)
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
		})
}

// handleHelmResourceWithCache : 캐시를 사용하는 Helm 리소스 처리 헬퍼
func (h *HelmHandler) handleHelmResourceWithCache(c echo.Context, cacheKey string, kubeConfig models.KubeConfig,
	getResource func(interfaces.HelmClient, context.Context) (interface{}, error)) error {

	// 캐시 키 생성
	fullCacheKey := fmt.Sprintf("helm:%s:%s", cacheKey, kubeConfig.Namespace)

	// 캐시에서 가져오기
	if cached, exists := h.cache.Get(fullCacheKey); exists {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   cached,
			"cached": true,
		})
	}

	// 캐시에 없으면 클라이언트에서 가져오기
	helmClient, err := client.NewHelmClient(kubeConfig)
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
