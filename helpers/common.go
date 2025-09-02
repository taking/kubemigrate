package helpers

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"taking.kr/velero/models"
	"taking.kr/velero/utils"
	"taking.kr/velero/validation"
	"time"
)

// BindAndValidateKubeConfig : 요청 바인딩 + kubeconfig 유효성 검사
func BindAndValidateKubeConfig(ctx echo.Context, validator *validation.RequestValidator) (models.KubeConfig, error) {
	var req models.KubeConfig
	if err := ctx.Bind(&req); err != nil {
		return req, utils.RespondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	decodeKubeConfig, err := validator.ValidateKubeConfigRequest(&req)
	if err != nil {
		return req, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	req.KubeConfig = decodeKubeConfig
	return req, nil
}

// ResolveNamespace : namespace 확인 (요청 > query > 기본값)
func ResolveNamespace(req *models.KubeConfig, ctx echo.Context, defaultNS string) string {
	if req.Namespace != "" {
		return req.Namespace
	}
	if ns := ctx.QueryParam("namespace"); ns != "" {
		return ns
	}
	return defaultNS
}

// TimeoutContext : 지정된 시간만큼 timeout context 생성
func TimeoutContext(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	if duration <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, duration)
}
