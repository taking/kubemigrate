package controller

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"taking.kr/velero/clients"
	"taking.kr/velero/models"
	"taking.kr/velero/utils"
)

// HelmController : Helm 관련 API 컨트롤러
type HelmController struct {
	*BaseController
}

func NewHelmController() *HelmController {
	return &HelmController{
		BaseController: NewBaseController(),
	}
}

// CheckHelmConnection : 헬름 연결 및 쿠버네티스 접근 확인
// CheckHelmConnection godoc
// @Summary 헬름 연결 확인
// @Description kubeconfig을 사용하여 헬름 연결 검증
// @Tags helm
// @Accept json
// @Produce json
// @Param request body models.HelmConfigRequest true "헬름 연결에 필요한 값"
// @Success 200 {object} models.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} models.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} models.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} models.SwaggerErrorResponse "서비스 이용 불가"
// @Router /helm/health [get]
func (c *HelmController) CheckHelmConnection(ctx echo.Context) error {
	// kubeConfig 바인딩 및 검증
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	// 네임스페이스 값이 없으면, 기본 네임스페이스 "default"로 설정
	req.Namespace = c.ResolveNamespace(&req, ctx, "default")

	// Helm 클라이언트 생성
	client, err := clients.NewHelmClient(req)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	_ = client.InvalidateCache() // 캐시 초기화 후 최신 정보로 동작

	// Helm 연결 상태 확인
	return c.HandleHealthCheck(ctx, client, "Helm")
}

// IsChartInstalled : 특정 헬름 차트 설치 여부 확인
// IsChartInstalled godoc
// @Summary 특정 헬름 차트 설치 여부 확인
// @Description kubeconfig을 사용하여 특정 헬름 차트 설치 여부 확인
// @Tags helm
// @Accept json
// @Produce json
// @Param request body models.HelmConfigRequest true "헬름 연결에 필요한 값"
// @Success 200 {object} models.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} models.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} models.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} models.SwaggerErrorResponse "서비스 이용 불가"
// @Router /helm/chart_check [post]
func (c *HelmController) IsChartInstalled(ctx echo.Context) error {
	// KubeConfig 바인딩 및 검증
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	// Helm 클라이언트 생성
	client, err := clients.NewHelmClient(req)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	_ = client.InvalidateCache() // 캐시 초기화 후 최신 정보로 동작

	// 차트 설치 여부 확인
	installed, data, err := client.IsChartInstalled(req.ChartName)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to check chart installation: %v", err))
	}

	return utils.RespondSuccess(ctx, map[string]interface{}{
		"chart":     data,
		"installed": installed,
	})
}

// InstallChart : 헬름 차트 설치
// InstallChart godoc
// @Summary 헬름 차트 설치
// @Description kubeconfig을 사용하여 헬름 차트 설치
// @Tags helm
// @Accept json
// @Produce json
// @Param request body models.HelmInstallChartRequest true "헬름 연결에 필요한 값"
// @Success 200 {object} models.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} models.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} models.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} models.SwaggerErrorResponse "서비스 이용 불가"
// @Router /helm/install [post]
func (c *HelmController) InstallChart(ctx echo.Context) error {
	var req struct {
		models.KubeConfig
		ChartName string                 `json:"chartName"`
		ChartPath string                 `json:"chartPath"`
		Values    map[string]interface{} `json:"values"`
	}

	if err := ctx.Bind(&req); err != nil {
		return utils.RespondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	// Helm 클라이언트 생성
	client, err := clients.NewHelmClient(req.KubeConfig)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	_ = client.InvalidateCache() // 캐시 초기화 후 최신 정보로 동작

	// 차트 설치
	if err := client.InstallChart(req.ChartName, req.ChartPath, req.Values); err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to install chart '%s': %v", req.ChartName, err))
	}

	return utils.RespondStatus(ctx, "success", fmt.Sprintf("chart installed successfully: %s", req.ChartName))
}
