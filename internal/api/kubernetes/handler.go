package kubernetes

import (
	"context"
	"sort"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	"github.com/taking/kubemigrate/pkg/utils"
)

// Handler : Kubernetes 관련 HTTP 핸들러
type Handler struct {
	*handler.BaseHandler
}

// NewHandler : 새로운 Kubernetes 핸들러 생성
func NewHandler(base *handler.BaseHandler) *Handler {
	return &Handler{
		BaseHandler: base,
	}
}

// GetPods : Pod 목록 조회
// @Summary Get Pods
// @Description Get list of pods in specified namespace
// @Tags kubernetes
// @Accept json
// @Produce json
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/kubernetes/pods [post]
func (h *Handler) GetPods(c echo.Context) error {
	return h.HandleKubernetesResource(c, "pods", func(k8sClient kubernetes.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		req, err := utils.BindAndValidateKubeConfig(c, h.KubernetesValidator)
		if err != nil {
			return nil, err
		}

		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(&req, c, "default")

		// Pod 목록 조회
		pods, err := k8sClient.GetPods(ctx, namespace)
		if err != nil {
			return nil, err
		}

		// 생성 시간 기준으로 정렬
		sort.Slice(pods.Items, func(i, j int) bool {
			return pods.Items[j].CreationTimestamp.Before(&pods.Items[i].CreationTimestamp)
		})

		return pods, nil
	})
}

// GetStorageClasses : StorageClass 목록 조회
// @Summary Get Storage Classes
// @Description Get list of storage classes
// @Tags kubernetes
// @Accept json
// @Produce json
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/kubernetes/storage-classes [post]
func (h *Handler) GetStorageClasses(c echo.Context) error {
	return h.HandleKubernetesResource(c, "storage-classes", func(k8sClient kubernetes.Client, ctx context.Context) (interface{}, error) {
		// StorageClass 목록 조회
		storageClasses, err := k8sClient.GetStorageClasses(ctx)
		if err != nil {
			return nil, err
		}

		// 이름 기준으로 정렬
		sort.Slice(storageClasses.Items, func(i, j int) bool {
			return storageClasses.Items[i].Name < storageClasses.Items[j].Name
		})

		return storageClasses, nil
	})
}

// HealthCheck : Kubernetes 연결 상태 확인
// @Summary Kubernetes Health Check
// @Description Check Kubernetes connection status
// @Tags kubernetes
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/kubernetes/health [get]
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.HandleKubernetesResource(c, "kubernetes-health", func(k8sClient kubernetes.Client, ctx context.Context) (interface{}, error) {
		// 간단한 Kubernetes 연결 테스트 (네임스페이스 목록 조회)
		_, err := k8sClient.GetPods(ctx, "default")
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"service": "kubernetes",
			"status":  "healthy",
			"message": "Kubernetes connection is working",
		}, nil
	})
}
