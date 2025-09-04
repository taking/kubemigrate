package usecase

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/taking/velero/internal/app"
	"github.com/taking/velero/internal/model"
	"github.com/taking/velero/pkg/response"
	"github.com/taking/velero/pkg/validator"
	"net/http"
)

// BaseController : 공통 기능을 제공하는 기본 컨트롤러 구조체
type BaseController struct {
	kubernetesValidator *validator.KubernetesValidator
	minioValidator      *validator.MinioValidator
}

// NewBaseController : BaseController 객체 생성
func NewBaseController() *BaseController {
	return &BaseController{
		kubernetesValidator: validator.NewKubernetesValidator(),
		minioValidator:      validator.NewMinioValidator(),
	}
}

// BindAndValidateKubeConfig : 요청으로 들어온 KubeConfig 데이터를 바인딩하고 검증
func (c *BaseController) BindAndValidateKubeConfig(ctx echo.Context) (model.KubeConfig, error) {
	return app.BindAndValidateKubeConfig(ctx, c.kubernetesValidator)
}

// BindAndValidateMinioConfig : 요청으로 들어온 MinioConfig 데이터를 바인딩하고 검증
func (c *BaseController) BindAndValidateMinioConfig(ctx echo.Context) (model.MinioConfig, error) {
	return app.BindAndValidateMinioConfig(ctx, c.minioValidator)
}

// BindAndValidateVeleroConfig : 요청으로 들어온 MinioConfig 및 KubeConfig 데이터를 바인딩하고 검증
func (c *BaseController) BindAndValidateVeleroConfig(ctx echo.Context) (model.VeleroConfig, error) {
	return app.BindAndValidateVeleroConfig(ctx, c.minioValidator, c.kubernetesValidator)
}

// ResolveNamespace : 네임스페이스 결정
func (c *BaseController) ResolveNamespace(req *model.KubeConfig, ctx echo.Context, defaultNS string) string {
	return app.ResolveNamespace(req, ctx, defaultNS)
}

// HandleHealthCheck : 공통 HealthCheck 핸들러
func (c *BaseController) HandleHealthCheck(ctx echo.Context, healthChecker interface {
	HealthCheck(context.Context) error
}, serviceName string) error {
	if err := healthChecker.HealthCheck(ctx.Request().Context()); err != nil {
		return response.RespondError(ctx, http.StatusServiceUnavailable,
			fmt.Sprintf("%s cluster unhealthy: %s", serviceName, err.Error()))
	}
	return response.RespondStatus(ctx, "healthy", fmt.Sprintf("%s connection successful", serviceName))
}
