package controller

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"taking.kr/velero/helpers"
	"taking.kr/velero/models"
	"taking.kr/velero/utils"
	"taking.kr/velero/validation"
)

// BaseController : 공통 기능을 제공하는 기본 컨트롤러 구조체
type BaseController struct {
	validator *validation.KubernetesValidator
}

// NewBaseController : BaseController 객체 생성
func NewBaseController() *BaseController {
	return &BaseController{
		validator: validation.NewKubernetesValidator(),
	}
}

// BindAndValidateKubeConfig : 요청으로 들어온 KubeConfig 데이터를 바인딩하고 검증
func (c *BaseController) BindAndValidateKubeConfig(ctx echo.Context) (models.KubeConfig, error) {
	return helpers.BindAndValidateKubeConfig(ctx, c.validator)
}

// ResolveNamespace : 네임스페이스 결정
func (c *BaseController) ResolveNamespace(req *models.KubeConfig, ctx echo.Context, defaultNS string) string {
	return helpers.ResolveNamespace(req, ctx, defaultNS)
}

// HandleHealthCheck : 공통 HealthCheck 핸들러
func (c *BaseController) HandleHealthCheck(ctx echo.Context, healthChecker interface {
	HealthCheck(context.Context) error
}, serviceName string) error {
	if err := healthChecker.HealthCheck(ctx.Request().Context()); err != nil {
		return utils.RespondError(ctx, http.StatusServiceUnavailable,
			fmt.Sprintf("%s cluster unhealthy: %s", serviceName, err.Error()))
	}
	return utils.RespondStatus(ctx, "healthy", fmt.Sprintf("%s connection successful", serviceName))
}
