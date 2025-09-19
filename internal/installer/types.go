package installer

import (
	"context"
	"time"

	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
)

// VeleroInstallConfig : Velero 설치 설정
type VeleroInstallConfig struct {
	MinioConfig config.MinioConfig `json:"minioConfig"`
	Namespace   string             `json:"namespace"`
	Force       bool               `json:"force"`
}

// VeleroUninstallConfig : Velero 제거 설정
type VeleroUninstallConfig struct {
	Namespace string `json:"namespace"`
	Force     bool   `json:"force"`
}

// InstallResult : 설치 결과
type InstallResult struct {
	Status           string                 `json:"status"`
	Message          string                 `json:"message"`
	VeleroNamespace  string                 `json:"veleroNamespace"`
	Force            bool                   `json:"force"`
	MinioConnected   bool                   `json:"minioConnected"`
	BackupLocation   string                 `json:"backupLocation"`
	InstallationTime time.Duration          `json:"installationTime"`
	Details          map[string]interface{} `json:"details"`
}

// VeleroStatus : Velero 상태 정보
type VeleroStatus struct {
	IsHealthy     bool `json:"isHealthy"`
	PodsInstalled bool `json:"podsInstalled"`
	HelmRelease   bool `json:"helmRelease"`
}

// InstallationStrategy : 설치 전략
type InstallationStrategy string

const (
	StrategyForceReinstall InstallationStrategy = "force_reinstall"
	StrategyFreshInstall   InstallationStrategy = "fresh_install"
	StrategySkipInstall    InstallationStrategy = "skip_install"
)

// InstallerService : 설치 서비스 인터페이스
type InstallerService interface {
	InstallVelero(ctx context.Context, client client.Client, config VeleroInstallConfig) (*InstallResult, error)
	UninstallVelero(ctx context.Context, client client.Client, config VeleroUninstallConfig) error
	CleanupVelero(ctx context.Context, client client.Client, namespace string, force bool) error
	DetermineStrategy(ctx context.Context, client client.Client, config VeleroInstallConfig) (InstallationStrategy, *VeleroStatus, error)
}
