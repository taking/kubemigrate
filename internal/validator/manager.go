package validator

import (
	"fmt"

	"github.com/taking/kubemigrate/pkg/config"
)

// ValidationManager : 통합 검증 관리자
type ValidationManager struct {
	kubernetesValidator *KubernetesValidator
	minioValidator      *MinioValidator
}

// NewValidationManager : 새로운 검증 관리자 생성
func NewValidationManager() *ValidationManager {
	return &ValidationManager{
		kubernetesValidator: NewKubernetesValidator(),
		minioValidator:      NewMinioValidator(),
	}
}

// GetKubernetesValidator : Kubernetes 검증자 반환
func (vm *ValidationManager) GetKubernetesValidator() *KubernetesValidator {
	return vm.kubernetesValidator
}

// GetMinioValidator : MinIO 검증자 반환
func (vm *ValidationManager) GetMinioValidator() *MinioValidator {
	return vm.minioValidator
}

// ValidateKubeConfig : Kubernetes 설정 검증
func (vm *ValidationManager) ValidateKubeConfig(kubeConfig *config.KubeConfig) error {
	_, err := vm.kubernetesValidator.ValidateKubernetesConfig(kubeConfig)
	return err
}

// ValidateMinioConfig : MinIO 설정 검증
func (vm *ValidationManager) ValidateMinioConfig(minioConfig *config.MinioConfig) error {
	return vm.minioValidator.ValidateMinioConfig(minioConfig)
}

// ValidateVeleroConfig : Velero 설정 검증
func (vm *ValidationManager) ValidateVeleroConfig(veleroConfig *config.VeleroConfig) error {
	// MinIO 설정 검증
	if err := vm.ValidateMinioConfig(&veleroConfig.MinioConfig); err != nil {
		return fmt.Errorf("minio config validation failed: %w", err)
	}

	// Kubernetes 설정 검증
	if err := vm.ValidateKubeConfig(&veleroConfig.KubeConfig); err != nil {
		return fmt.Errorf("kubernetes config validation failed: %w", err)
	}

	return nil
}

// ValidateInstallChartConfig : Helm 차트 설치 설정 검증
func (vm *ValidationManager) ValidateInstallChartConfig(installConfig *config.InstallChartConfig) error {
	if installConfig.ReleaseName == "" {
		return fmt.Errorf("release name is required")
	}

	if installConfig.ChartURL == "" {
		return fmt.Errorf("chart URL is required")
	}

	if installConfig.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	return nil
}

// ValidateUpgradeChartConfig : Helm 차트 업그레이드 설정 검증
func (vm *ValidationManager) ValidateUpgradeChartConfig(upgradeConfig *config.UpgradeChartConfig) error {
	if upgradeConfig.ReleaseName == "" {
		return fmt.Errorf("release name is required")
	}

	if upgradeConfig.ChartPath == "" {
		return fmt.Errorf("chart path is required")
	}

	if upgradeConfig.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	return nil
}

// ValidateAll : 모든 설정 검증
func (vm *ValidationManager) ValidateAll(configs map[string]interface{}) error {
	var errors []error

	for configType, configValue := range configs {
		switch configType {
		case "kubeconfig":
			if kubeConfig, ok := configValue.(*config.KubeConfig); ok {
				if err := vm.ValidateKubeConfig(kubeConfig); err != nil {
					errors = append(errors, fmt.Errorf("kubernetes config: %w", err))
				}
			}
		case "minio":
			if minioConfig, ok := configValue.(*config.MinioConfig); ok {
				if err := vm.ValidateMinioConfig(minioConfig); err != nil {
					errors = append(errors, fmt.Errorf("minio config: %w", err))
				}
			}
		case "velero":
			if veleroConfig, ok := configValue.(*config.VeleroConfig); ok {
				if err := vm.ValidateVeleroConfig(veleroConfig); err != nil {
					errors = append(errors, fmt.Errorf("velero config: %w", err))
				}
			}
		case "install_chart":
			if installConfig, ok := configValue.(*config.InstallChartConfig); ok {
				if err := vm.ValidateInstallChartConfig(installConfig); err != nil {
					errors = append(errors, fmt.Errorf("install chart config: %w", err))
				}
			}
		case "upgrade_chart":
			if upgradeConfig, ok := configValue.(*config.UpgradeChartConfig); ok {
				if err := vm.ValidateUpgradeChartConfig(upgradeConfig); err != nil {
					errors = append(errors, fmt.Errorf("upgrade chart config: %w", err))
				}
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	return nil
}
