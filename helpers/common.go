package helpers

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"taking.kr/velero/models"
	"taking.kr/velero/utils"
	"taking.kr/velero/validation"
)

// BindAndValidateKubeConfig : 요청 바인딩 + kubeconfig 유효성 검사
func BindAndValidateKubeConfig(ctx echo.Context, validator *validation.KubernetesValidator) (models.KubeConfig, error) {
	var req models.KubeConfig
	if err := ctx.Bind(&req); err != nil {
		return req, utils.RespondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	decodeKubeConfig, err := validator.ValidateKubernetesConfig(&req)
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

// RunWithTimeout : 지정한 함수(func() error)를 주어진 context와 함께 실행하고, context가 만료되면 timeout 에러를 반환
func RunWithTimeout(ctx context.Context, fn func() error) error {
	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("operation failed: timeout")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}
		return nil
	}
}
