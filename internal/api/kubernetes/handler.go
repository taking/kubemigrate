package kubernetes

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client"
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

// HealthCheck : Kubernetes 연결 테스트
// @Summary Kubernetes Connection Test
// @Description Test Kubernetes connection with provided configuration
// @Tags kubernetes
// @Accept json
// @Produce json
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/kubernetes/health [post]
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.HandleResourceClient(c, "kubernetes-health", func(client client.Client, ctx context.Context) (interface{}, error) {
		// Kubernetes 연결 테스트
		_, err := client.Kubernetes().GetPods(ctx, "default", "")
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"service": "kubernetes",
			"message": "Kubernetes connection is working",
		}, nil
	})
}

// GetResources : Kubernetes 리소스 조회
// @Summary Get Kubernetes Resources
// @Description Get Kubernetes resources by kind, name (optional) and namespace
// @Tags kubernetes
// @Accept json
// @Produce json
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param kind path string true "Resource kind (pods, configmaps, secrets, storage-classes)"
// @Param name path string false "Resource name (empty for list, specific name for single resource)"
// @Param namespace query string false "Namespace name (default: 'default', all namespaces: 'all')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/kubernetes/{kind}/{name} [get]
func (h *Handler) GetResources(c echo.Context) error {
	return h.HandleResourceClient(c, "resources", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		// "all"이면 모든 네임스페이스 조회,""이면 3번째 파라미터 값을 네임스페이스로 사용
		namespace := utils.ResolveNamespace(c, "default")

		// GET 요청에서는 body 바인딩 없이 query parameter만 사용
		// 리소스 종류, 이름, 네임스페이스 결정
		kind := c.Param("kind")
		name := c.Param("name")

		// 클라이언트 통합 메서드 사용
		switch kind {
		case "pods":
			return client.Kubernetes().GetPods(ctx, namespace, name)
		case "configmaps":
			return client.Kubernetes().GetConfigMaps(ctx, namespace, name)
		case "secrets":
			return client.Kubernetes().GetSecrets(ctx, namespace, name)
		case "storage-classes":
			return client.Kubernetes().GetStorageClasses(ctx, name)
		default:
			supportedResources := []string{"pods", "configmaps", "secrets", "storage-classes"}
			errorMsg := fmt.Sprintf("Unsupported resource kind: %s. Supported resources: %v", kind, supportedResources)
			return nil, echo.NewHTTPError(400, errorMsg)
		}
	})
}
