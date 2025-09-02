package controller

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"taking.kr/velero/clients"
	"taking.kr/velero/helpers"
	"taking.kr/velero/models"
	"taking.kr/velero/validation"
)

// KubernetesController : K8s 관련 API 컨트롤러
type KubernetesController struct {
	validator *validation.RequestValidator
}

func NewKubernetesController() *KubernetesController {
	return &KubernetesController{
		validator: validation.NewRequestValidator(),
	}
}

// CheckKubernetesConnection : Kubernetes 클러스터 연결 확인
func (c *KubernetesController) CheckKubernetesConnection(ctx echo.Context) error {
	req, err := helpers.BindAndValidateKubeConfig(ctx, c.validator)
	if err != nil {
		return err
	}

	namespace := helpers.ResolveNamespace(&req, ctx, "default")

	client, err := clients.NewKubeClient(models.KubeConfig{
		KubeConfig: req.KubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	if err := client.HealthCheck(ctx.Request().Context()); err != nil {
		return helpers.JSONError(ctx, http.StatusServiceUnavailable, "Kubernetes cluster unhealthy: "+err.Error())
	}

	return helpers.JSONStatus(ctx, "healthy", "Kubernetes cluster connection successful")
}

// GetPods : 네임스페이스 내 모든 Pod 조회
func (c *KubernetesController) GetPods(ctx echo.Context) error {
	req, err := helpers.BindAndValidateKubeConfig(ctx, c.validator)
	if err != nil {
		return err
	}

	client, err := clients.NewKubeClient(req)
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	data, err := client.GetPods(context.Background())
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	return helpers.JSONSuccess(ctx, data)
}

// GetStorageClasses : 클러스터 내 StorageClass 조회
func (c *KubernetesController) GetStorageClasses(ctx echo.Context) error {
	req, err := helpers.BindAndValidateKubeConfig(ctx, c.validator)
	if err != nil {
		return err
	}

	client, err := clients.NewKubeClient(req)
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	data, err := client.GetStorageClasses(context.Background())
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	return helpers.JSONSuccess(ctx, data)
}
