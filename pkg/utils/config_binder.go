package utils

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/internal/validator"
	"github.com/taking/kubemigrate/pkg/config"
)

// ConfigBinder : 통합 설정 바인딩 유틸리티
type ConfigBinder struct {
	validationManager *validator.ValidationManager
}

// NewConfigBinder : 새로운 설정 바인딩 유틸리티 생성
func NewConfigBinder() *ConfigBinder {
	return &ConfigBinder{
		validationManager: validator.NewValidationManager(),
	}
}

// BindKubeConfig : Kubernetes 설정 바인딩 및 검증
func (cb *ConfigBinder) BindKubeConfig(c echo.Context) (config.KubeConfig, error) {
	var kubeConfig config.KubeConfig
	if err := c.Bind(&kubeConfig); err != nil {
		return kubeConfig, response.RespondError(c, 400, "invalid request body")
	}

	if err := cb.validationManager.ValidateKubeConfig(&kubeConfig); err != nil {
		return kubeConfig, echo.NewHTTPError(400, err.Error())
	}

	return kubeConfig, nil
}

// BindMinioConfig : MinIO 설정 바인딩 및 검증
func (cb *ConfigBinder) BindMinioConfig(c echo.Context) (config.MinioConfig, error) {
	var minioConfig config.MinioConfig
	if err := c.Bind(&minioConfig); err != nil {
		return minioConfig, response.RespondError(c, 400, "invalid request body")
	}

	if err := cb.validationManager.ValidateMinioConfig(&minioConfig); err != nil {
		return minioConfig, fmt.Errorf("minio config validation failed: %w", err)
	}

	return minioConfig, nil
}

// BindVeleroConfig : Velero 설정 바인딩 및 검증
func (cb *ConfigBinder) BindVeleroConfig(c echo.Context) (config.VeleroConfig, error) {
	var veleroConfig config.VeleroConfig
	if err := c.Bind(&veleroConfig); err != nil {
		return veleroConfig, response.RespondError(c, 400, "invalid request body")
	}

	if err := cb.validationManager.ValidateVeleroConfig(&veleroConfig); err != nil {
		return veleroConfig, fmt.Errorf("velero config validation failed: %w", err)
	}

	return veleroConfig, nil
}

// BindInstallChartConfig : Helm 차트 설치 설정 바인딩 및 검증
func (cb *ConfigBinder) BindInstallChartConfig(c echo.Context) (config.InstallChartConfig, error) {
	var installConfig config.InstallChartConfig
	if err := c.Bind(&installConfig); err != nil {
		return installConfig, response.RespondError(c, 400, "invalid request body")
	}

	if err := cb.validationManager.ValidateInstallChartConfig(&installConfig); err != nil {
		return installConfig, fmt.Errorf("install chart config validation failed: %w", err)
	}

	return installConfig, nil
}

// BindUpgradeChartConfig : Helm 차트 업그레이드 설정 바인딩 및 검증
func (cb *ConfigBinder) BindUpgradeChartConfig(c echo.Context) (config.UpgradeChartConfig, error) {
	var upgradeConfig config.UpgradeChartConfig
	if err := c.Bind(&upgradeConfig); err != nil {
		return upgradeConfig, response.RespondError(c, 400, "invalid request body")
	}

	if err := cb.validationManager.ValidateUpgradeChartConfig(&upgradeConfig); err != nil {
		return upgradeConfig, fmt.Errorf("upgrade chart config validation failed: %w", err)
	}

	return upgradeConfig, nil
}

// BindAndValidate : 설정 타입에 따른 바인딩 및 검증
func (cb *ConfigBinder) BindAndValidate(c echo.Context, configType string) (interface{}, error) {
	switch configType {
	case "kubeconfig":
		return cb.BindKubeConfig(c)
	case "minio":
		return cb.BindMinioConfig(c)
	case "velero":
		return cb.BindVeleroConfig(c)
	case "install_chart":
		return cb.BindInstallChartConfig(c)
	case "upgrade_chart":
		return cb.BindUpgradeChartConfig(c)
	default:
		return nil, fmt.Errorf("unsupported config type: %s", configType)
	}
}

// GetValidationManager : 검증 관리자 반환
func (cb *ConfigBinder) GetValidationManager() *validator.ValidationManager {
	return cb.validationManager
}
