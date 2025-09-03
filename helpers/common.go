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

// BindAndValidateMinioConfig : 요청 바인딩 + minioConfig 유효성 검사
func BindAndValidateMinioConfig(ctx echo.Context, minioValidator *validation.MinioValidator) (models.MinioConfig, error) {
	var req models.MinioConfig
	if err := ctx.Bind(&req); err != nil {
		return req, utils.RespondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	// minio config validator
	if err := minioValidator.ValidateMinioConfig(&req); err != nil {
		return req, fmt.Errorf("minio config validation failed: %w", err)
	}
	return req, nil
}

// BindAndValidateVeleroConfig : minioConfig, kubeConfig 요청 바인딩 + kubeConfig 유효성 검사
func BindAndValidateVeleroConfig(ctx echo.Context, minioValidator *validation.MinioValidator, kubernetesValidator *validation.KubernetesValidator) (models.VeleroConfig, error) {
	var req models.VeleroConfig
	if err := ctx.Bind(&req); err != nil {
		return req, utils.RespondError(ctx, http.StatusBadRequest, "invalid request body")
	}

	// minio config validator
	if err := minioValidator.ValidateMinioConfig(&req.MinioConfig); err != nil {
		return req, fmt.Errorf("minio config validation failed: %w", err)
	}

	// Kubernetes config validator & base64 Decode
	decodeKubeConfig, err := kubernetesValidator.ValidateKubernetesConfig(&req.KubeConfig)
	if err != nil {
		return req, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	req.KubeConfig.KubeConfig = decodeKubeConfig
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
