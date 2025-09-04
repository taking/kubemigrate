package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"taking.kr/velero/pkg/cache"
	"taking.kr/velero/pkg/client"
	"taking.kr/velero/pkg/health"
	"taking.kr/velero/pkg/interfaces"
	"taking.kr/velero/pkg/models"
	"taking.kr/velero/pkg/response"
	"taking.kr/velero/pkg/utils"
	"taking.kr/velero/pkg/validator"
)

// KubernetesHandler : Kubernetes 관련 HTTP 요청을 처리하는 핸들러
type KubernetesHandler struct {
	kubernetesValidator *validator.KubernetesValidator
	cache               *cache.Cache
	workerPool          *utils.WorkerPool
	healthManager       *health.HealthManager
}

// NewKubernetesHandler : 새로운 KubernetesHandler 인스턴스 생성
func NewKubernetesHandler(appCache *cache.Cache, workerPool *utils.WorkerPool, healthManager *health.HealthManager) *KubernetesHandler {
	return &KubernetesHandler{
		kubernetesValidator: validator.NewKubernetesValidator(),
		cache:               appCache,
		workerPool:          workerPool,
		healthManager:       healthManager,
	}
}

// HealthCheck : Kubernetes 연결 상태 확인
// @Summary Kubernetes 연결 상태 확인
// @Description Validate Kubernetes connection using KubeConfig
// @Tags kubernetes
// @Accept json
// @Produce json
// @Param request body models.KubeConfig true "Kubernetes connection configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Connection successful"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Failure 503 {object} models.SwaggerErrorResponse "Service unavailable"
// @Router /kube/health [get]
func (h *KubernetesHandler) HealthCheck(c echo.Context) error {
	req, err := h.bindAndValidateKubeConfig(c)
	if err != nil {
		return err
	}

	// 기본 네임스페이스 설정
	req.Namespace = resolveNamespace(&req, c, "default")

	client, err := client.NewKubernetesClient(req)
	if err != nil {
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}

	// 클러스터 연결 상태 확인
	if err := client.HealthCheck(c.Request().Context()); err != nil {
		return response.RespondError(c, http.StatusServiceUnavailable,
			"Kubernetes cluster unhealthy: "+err.Error())
	}

	// 헬스체크 매니저에 Kubernetes 체커 등록
	if h.healthManager != nil {
		kubeChecker := health.NewKubernetesHealthChecker(req)
		h.healthManager.Register(kubeChecker)
	}

	return response.RespondStatus(c, "healthy", "Kubernetes connection successful")
}

// GetPods : Kubernetes pod 목록 조회
// @Summary Kubernetes pod 목록 조회
// @Description Retrieve pod list from Kubernetes using KubeConfig
// @Tags kubernetes
// @Accept json
// @Produce json
// @Param request body models.KubeConfig true "Kubernetes connection configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Success"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /kube/pods [get]
func (h *KubernetesHandler) GetPods(c echo.Context) error {
	return h.handleKubernetesResourceWithCache(c, "pods", func(client interfaces.KubernetesClient, ctx context.Context) (interface{}, error) {
		return client.GetPods(ctx)
	})
}

// GetStorageClasses : Kubernetes storage class 목록 조회
// @Summary Kubernetes storage class 목록 조회
// @Description Retrieve storage class list from Kubernetes using KubeConfig
// @Tags kubernetes
// @Accept json
// @Produce json
// @Param request body models.KubeConfig true "Kubernetes connection configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Success"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /kube/storage-classes [get]
func (h *KubernetesHandler) GetStorageClasses(c echo.Context) error {
	return h.handleKubernetesResourceWithCache(c, "storage-classes", func(client interfaces.KubernetesClient, ctx context.Context) (interface{}, error) {
		return client.GetStorageClasses(ctx)
	})
}

// handleKubernetesResourceWithCache : 캐시를 사용하는 Kubernetes 리소스 처리 헬퍼
func (h *KubernetesHandler) handleKubernetesResourceWithCache(c echo.Context, cacheKey string,
	getResource func(interfaces.KubernetesClient, context.Context) (interface{}, error)) error {

	// KubeConfig 검증
	req, err := h.bindAndValidateKubeConfig(c)
	if err != nil {
		return err
	}

	// 기본 네임스페이스 설정
	req.Namespace = resolveNamespace(&req, c, "default")

	// 캐시 키 생성
	fullCacheKey := fmt.Sprintf("kubernetes:%s:%s", cacheKey, req.Namespace)

	// 캐시에서 가져오기
	if cached, exists := h.cache.Get(fullCacheKey); exists {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   cached,
			"cached": true,
		})
	}

	// 캐시에 없으면 클라이언트에서 가져오기
	kubernetesClient, err := client.NewKubernetesClient(req)
	if err != nil {
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}

	// 워커 풀을 사용하여 백그라운드에서 데이터 가져오기
	resultChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)

	h.workerPool.Submit(func() {
		data, err := getResource(kubernetesClient, c.Request().Context())
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

// bindAndValidateKubeConfig : KubeConfig 검증
func (h *KubernetesHandler) bindAndValidateKubeConfig(c echo.Context) (models.KubeConfig, error) {
	var req models.KubeConfig
	if err := c.Bind(&req); err != nil {
		return req, response.RespondError(c, http.StatusBadRequest, "invalid request body")
	}

	decodedKubeConfig, err := h.kubernetesValidator.ValidateKubernetesConfig(&req)
	if err != nil {
		return req, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	req.KubeConfig = decodedKubeConfig
	return req, nil
}
