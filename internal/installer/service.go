package installer

import (
	"context"

	"github.com/taking/kubemigrate/pkg/client"
)

// Service : 통합 설치 서비스
type Service struct {
	veleroInstaller *VeleroInstaller
}

// NewService : 새로운 설치 서비스 생성
func NewService() *Service {
	return &Service{
		veleroInstaller: NewVeleroInstaller(),
	}
}

// InstallVelero : Velero 설치
func (s *Service) InstallVelero(ctx context.Context, client client.Client, config VeleroInstallConfig) (*InstallResult, error) {
	return s.veleroInstaller.Install(ctx, client, config)
}

// UninstallVelero : Velero 제거
func (s *Service) UninstallVelero(ctx context.Context, client client.Client, config VeleroUninstallConfig) error {
	return s.veleroInstaller.Uninstall(ctx, client, config)
}

// CleanupVelero : Velero 정리
func (s *Service) CleanupVelero(ctx context.Context, client client.Client, namespace string, force bool) error {
	return s.veleroInstaller.Cleanup(ctx, client, namespace, force)
}

// DetermineStrategy : 설치 전략 결정
func (s *Service) DetermineStrategy(ctx context.Context, client client.Client, config VeleroInstallConfig) (InstallationStrategy, *VeleroStatus, error) {
	return s.veleroInstaller.DetermineStrategy(ctx, client, config)
}
