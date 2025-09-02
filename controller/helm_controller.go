package controller

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"taking.kr/velero/clients"
	"taking.kr/velero/helpers"
	"taking.kr/velero/models"
	"taking.kr/velero/utils"
)

type HelmController struct {
	*BaseController
}

func NewHelmController() *HelmController {
	return &HelmController{
		BaseController: NewBaseController(),
	}
}

// CheckHelmConnection : Helm 연결 및 Kubernetes 접근 확인
func (c *HelmController) CheckHelmConnection(ctx echo.Context) error {
	// KubeConfig 바인딩 및 검증
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	// 기본 네임스페이스 설정
	req.Namespace = helpers.ResolveNamespace(&req, ctx, "default")

	// Helm 클라이언트 생성
	client, err := clients.NewHelmClient(req)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	_ = client.InvalidateCache() // 캐시 초기화 후 최신 정보로 동작

	// Helm 연결 상태 확인
	return c.HandleHealthCheck(ctx, client, "Helm")
}

// IsChartInstalled : 특정 Helm 차트 설치 여부 확인
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

// InstallChart : Helm 차트 설치
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
