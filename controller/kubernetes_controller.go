package controller

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"taking.kr/velero/clients"
	"taking.kr/velero/utils"
)

// KubernetesController : K8s 관련 API 컨트롤러
type KubernetesController struct {
	*BaseController
}

func NewKubernetesController() *KubernetesController {
	return &KubernetesController{
		BaseController: NewBaseController(),
	}
}

// CheckKubernetesConnection : Kubernetes 클러스터 연결 확인
func (c *KubernetesController) CheckKubernetesConnection(ctx echo.Context) error {
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	namespace := c.ResolveNamespace(&req, ctx, "default")
	req.Namespace = namespace

	client, err := clients.NewKubeClient(req)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	return c.HandleHealthCheck(ctx, client, "Kubernetes")
}

// GetPods : 네임스페이스 내 모든 Pod 조회
func (c *KubernetesController) GetPods(ctx echo.Context) error {
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	client, err := clients.NewKubeClient(req)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	data, err := client.GetPods(context.Background())
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	return utils.RespondSuccess(ctx, data)
}

// GetStorageClasses : 클러스터 내 StorageClass 조회
func (c *KubernetesController) GetStorageClasses(ctx echo.Context) error {
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	client, err := clients.NewKubeClient(req)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	data, err := client.GetStorageClasses(context.Background())
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	return utils.RespondSuccess(ctx, data)
}
