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

// BaseController provides common controller functionality
type BaseController struct {
	validator *validation.RequestValidator
}

func NewBaseController() *BaseController {
	return &BaseController{
		validator: validation.NewRequestValidator(),
	}
}

// BindAndValidateKubeConfig binds and validates KubeConfig request
func (c *BaseController) BindAndValidateKubeConfig(ctx echo.Context) (models.KubeConfig, error) {
	return helpers.BindAndValidateKubeConfig(ctx, c.validator)
}

// ResolveNamespace resolves namespace with query param fallback
func (c *BaseController) ResolveNamespace(req *models.KubeConfig, ctx echo.Context, defaultNS string) string {
	return helpers.ResolveNamespace(req, ctx, defaultNS)
}

// HandleHealthCheck generic health check handler
func (c *BaseController) HandleHealthCheck(ctx echo.Context, healthChecker interface {
	HealthCheck(context.Context) error
}, serviceName string) error {
	if err := healthChecker.HealthCheck(ctx.Request().Context()); err != nil {
		return utils.RespondError(ctx, http.StatusServiceUnavailable,
			fmt.Sprintf("%s cluster unhealthy: %s", serviceName, err.Error()))
	}
	return utils.RespondStatus(ctx, "healthy", fmt.Sprintf("%s connection successful", serviceName))
}
