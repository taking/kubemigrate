package utils

import (
	"context"
	"fmt"
	"os"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/internal/validator"
)

// StripManagedFields 리소스의 metadata.managedFields를 제거
func StripManagedFields(obj metav1.Object) {
	obj.SetManagedFields(nil)
}

// GetStringOrDefault : value가 비어있지 않으면 value를 반환, 비어있으면 def를 반환
func GetStringOrDefault(value, def string) string {
	if value == "" {
		return def
	}
	return value
}

// GetBoolOrDefault : value를 반환 (bool은 항상 값이 있으므로 단순히 value 반환)
func GetBoolOrDefault(value bool, def bool) bool {
	return value
}

// 간단한 파일 복사 함수
func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

// StringToIntOrDefault : string을 int로 변환, 실패하면 기본값 반환
func StringToIntOrDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// StringToBoolOrDefault : string을 bool로 변환, 실패하면 기본값 반환
func StringToBoolOrDefault(s string, def bool) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return def
	}
	return b
}

// ResolveNamespace : 네임스페이스 결정
func ResolveNamespace(ctx echo.Context, defaultNS string) string {
	var namespace string

	if ns := ctx.QueryParam("namespace"); ns != "" {
		namespace = ns
	} else {
		return defaultNS
	}

	// "all"을 빈 문자열로 변환 (모든 namespace 조회)
	if namespace == "all" {
		return ""
	}

	return namespace
}

// RunWithTimeout : 타임아웃과 함께 함수 실행
func RunWithTimeout(ctx context.Context, fn func() error) error {
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// BindAndValidateKubeConfig : KubeConfig 검증
func BindAndValidateKubeConfig(ctx echo.Context, validator *validator.KubernetesValidator) (config.KubeConfig, error) {
	var req config.KubeConfig
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
func BindAndValidateMinioConfig(ctx echo.Context, minioValidator *validator.MinioValidator) (config.MinioConfig, error) {
	var req config.MinioConfig
	if err := ctx.Bind(&req); err != nil {
		return req, response.RespondError(ctx, 400, "invalid request body")
	}

	if err := minioValidator.ValidateMinioConfig(&req); err != nil {
		return req, fmt.Errorf("minio config validation failed: %w", err)
	}
	return req, nil
}

// BindAndValidateHelmConfig : HelmConfig 검증
func BindAndValidateHelmConfig(ctx echo.Context, kubernetesValidator *validator.KubernetesValidator) (config.HelmConfig, error) {
	var req config.HelmConfig
	if err := ctx.Bind(&req); err != nil {
		return req, response.RespondError(ctx, 400, "invalid request body")
	}

	decodeKubeConfig, err := kubernetesValidator.ValidateKubernetesConfig(&req.KubeConfig)
	if err != nil {
		return req, echo.NewHTTPError(400, err.Error())
	}

	req.KubeConfig.KubeConfig = decodeKubeConfig
	return req, nil
}

// BindAndValidateVeleroConfig : VeleroConfig 검증
func BindAndValidateVeleroConfig(ctx echo.Context, minioValidator *validator.MinioValidator, kubernetesValidator *validator.KubernetesValidator) (config.VeleroConfig, error) {
	var req config.VeleroConfig
	if err := ctx.Bind(&req); err != nil {
		return req, response.RespondError(ctx, 400, "invalid request body")
	}

	if err := minioValidator.ValidateMinioConfig(&req.MinioConfig); err != nil {
		return req, fmt.Errorf("minio config validation failed: %w", err)
	}

	decodeKubeConfig, err := kubernetesValidator.ValidateKubernetesConfig(&req.KubeConfig)
	if err != nil {
		return req, echo.NewHTTPError(400, err.Error())
	}

	req.KubeConfig.KubeConfig = decodeKubeConfig
	return req, nil
}
