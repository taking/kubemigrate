package utils

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"taking.kr/velero/pkg/models"
	"taking.kr/velero/pkg/response"
	"taking.kr/velero/pkg/validator"
)

// BindAndValidateKubeConfig : KubeConfig 검증
func BindAndValidateKubeConfig(ctx echo.Context, validator *validator.KubernetesValidator) (models.KubeConfig, error) {
	var req models.KubeConfig
	if err := ctx.Bind(&req); err != nil {
		return req, response.RespondError(ctx, 400, "invalid request body")
	}

	decodeKubeConfig, err := validator.ValidateKubernetesConfig(&req)
	if err != nil {
		return req, echo.NewHTTPError(400, err.Error())
	}

	req.KubeConfig = decodeKubeConfig
	return req, nil
}

// BindAndValidateMinioConfig : MinioConfig 검증
func BindAndValidateMinioConfig(ctx echo.Context, minioValidator *validator.MinioValidator) (models.MinioConfig, error) {
	var req models.MinioConfig
	if err := ctx.Bind(&req); err != nil {
		return req, response.RespondError(ctx, 400, "invalid request body")
	}

	// minio config validator
	if err := minioValidator.ValidateMinioConfig(&req); err != nil {
		return req, fmt.Errorf("minio config validation failed: %w", err)
	}
	return req, nil
}

// BindAndValidateVeleroConfig : VeleroConfig 검증
func BindAndValidateVeleroConfig(ctx echo.Context, minioValidator *validator.MinioValidator, kubernetesValidator *validator.KubernetesValidator) (models.VeleroConfig, error) {
	var req models.VeleroConfig
	if err := ctx.Bind(&req); err != nil {
		return req, response.RespondError(ctx, 400, "invalid request body")
	}

	// minio config validator
	if err := minioValidator.ValidateMinioConfig(&req.MinioConfig); err != nil {
		return req, fmt.Errorf("minio config validation failed: %w", err)
	}

	// Kubernetes config validator & base64 Decode
	decodeKubeConfig, err := kubernetesValidator.ValidateKubernetesConfig(&req.KubeConfig)
	if err != nil {
		return req, echo.NewHTTPError(400, err.Error())
	}

	req.KubeConfig.KubeConfig = decodeKubeConfig
	return req, nil
}

// ResolveNamespace : 네임스페이스 결정
func ResolveNamespace(req *models.KubeConfig, ctx echo.Context, defaultNS string) string {
	if req.Namespace != "" {
		return req.Namespace
	}
	if ns := ctx.QueryParam("namespace"); ns != "" {
		return ns
	}
	return defaultNS
}

// RunWithTimeout : 컨텍스트 타임아웃과 함께 함수 실행
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
