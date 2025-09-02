package controller

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"taking.kr/velero/clients"
	"taking.kr/velero/services"
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

// CheckKubernetesConnection : 쿠버네티스 클러스터 연결 확인
// CheckKubernetesConnection godoc
// @Summary 쿠버네티스 클러스터 연결 확인
// @Description kubeconfig을 사용하여 쿠버네티스 클러스터 연결 검증
// @Tags kubernetes
// @Accept json
// @Produce json
// @Param request body models.KubeConfigRequest true "쿠버네티스 연결에 필요한 값"
// @Success 200 {object} models.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} models.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} models.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} models.SwaggerErrorResponse "서비스 이용 불가"
// @Router /kube/health [get]
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
// GetPods godoc
// @Summary 쿠버네티스 파드 목록 조회
// @Description kubeconfig을 사용하여 쿠버네티스 파드 목록 조회
// @Tags kubernetes
// @Accept json
// @Produce json
// @Param request body models.KubeConfigRequest true "쿠버네티스 연결에 필요한 값"
// @Success 200 {object} models.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} models.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} models.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} models.SwaggerErrorResponse "서비스 이용 불가"
// @Router /kube/pods [get]
func (c *KubernetesController) GetPods(ctx echo.Context) error {
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	client, err := clients.NewKubeClient(req)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	service := services.NewKubernetesService(client)
	data, err := service.GetPods(ctx.Request().Context())
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

func (c *KubernetesController) GetPodsAndStorage(ctx echo.Context) error {
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	client, err := clients.NewKubeClient(req)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	service := services.NewKubernetesService(client)
	data, err := service.GetPodsWithStorageClasses(ctx.Request().Context())
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	return utils.RespondSuccess(ctx, data)
}
