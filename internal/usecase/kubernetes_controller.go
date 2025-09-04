package usecase

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/taking/velero/internal/client"
	"github.com/taking/velero/pkg/response"
	"net/http"
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
// @Param request body github.com/taking/velero/internal/model.KubeConfigRequest true "쿠버네티스 연결에 필요한 값"
// @Success 200 {object} model.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} model.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} model.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} model.SwaggerErrorResponse "서비스 이용 불가"
// @Router /kube/health [get]
func (c *KubernetesController) CheckKubernetesConnection(ctx echo.Context) error {
	// KubeConfig 바인딩 및 검증
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	// 네임스페이스 값이 없으면, 기본 네임스페이스 "default"로 설정
	req.Namespace = c.ResolveNamespace(&req, ctx, "default")

	// Kubernetes 클라이언트 생성
	client, err := client.NewKubeClient(req)
	if err != nil {
		return response.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	// Kubernetes 연결 상태 확인
	return c.HandleHealthCheck(ctx, client, "Kubernetes")
}

// GetPods : 네임스페이스 내 모든 Pod 조회
// GetPods godoc
// @Summary 쿠버네티스 파드 목록 조회
// @Description kubeconfig을 사용하여 쿠버네티스 파드 목록 조회
// @Tags kubernetes
// @Accept json
// @Produce json
// @Param request body github.com/taking/velero/internal/model.KubeConfigRequest true "쿠버네티스 연결에 필요한 값"
// @Success 200 {object} model.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} model.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} model.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} model.SwaggerErrorResponse "서비스 이용 불가"
// @Router /kube/pods [get]
func (c *KubernetesController) GetPods(ctx echo.Context) error {
	// KubeConfig 바인딩 및 검증
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	// Kubernetes 클라이언트 생성
	client, err := client.NewKubeClient(req)
	if err != nil {
		return response.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	// 파드 목록 조회
	data, err := client.GetPods(context.Background())
	if err != nil {
		return response.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.RespondSuccess(ctx, data)
}

// GetStorageClasses : 클러스터 내 StorageClass 조회
func (c *KubernetesController) GetStorageClasses(ctx echo.Context) error {
	// KubeConfig 바인딩 및 검증
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	// Kubernetes 클라이언트 생성
	client, err := client.NewKubeClient(req)
	if err != nil {
		return response.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	// 스토리지 클래스 목록 조회
	data, err := client.GetStorageClasses(context.Background())
	if err != nil {
		return response.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.RespondSuccess(ctx, data)
}
