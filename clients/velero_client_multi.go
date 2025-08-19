package clients

import (
	"context"
	"fmt"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	velerov2 "github.com/vmware-tanzu/velero/pkg/apis/velero/v2alpha1"
	"taking.kr/velero/interfaces"
)

type multiClusterVeleroClient struct {
	sourceClient      interfaces.VeleroService
	destinationClient interfaces.VeleroService
	config            *interfaces.MultiClusterConfig
}

func NewMultiClusterVeleroClient(config *interfaces.MultiClusterConfig) (interfaces.MultiClusterVeleroService, error) {
	var sourceClient, destClient interfaces.VeleroService
	var err error

	if config.SourceKubeconfig != "" {
		sourceClient, err = NewVeleroClientFromRawConfig(config.SourceKubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create source client: %w", err)
		}
	}

	if config.DestinationKubeconfig != "" {
		destClient, err = NewVeleroClientFromRawConfig(config.DestinationKubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create destination client: %w", err)
		}
	}

	return &multiClusterVeleroClient{
		sourceClient:      sourceClient,
		destinationClient: destClient,
		config:            config,
	}, nil
}

func (m *multiClusterVeleroClient) GetBackups(ctx context.Context, namespace string) ([]velerov1.Backup, error) {
	if m.sourceClient == nil {
		return nil, fmt.Errorf("source client not configured")
	}
	return m.sourceClient.GetBackups(ctx, namespace)
}

func (m *multiClusterVeleroClient) GetRestores(ctx context.Context, namespace string) ([]velerov1.Restore, error) {
	if m.sourceClient == nil {
		return nil, fmt.Errorf("source client not configured")
	}
	return m.sourceClient.GetRestores(ctx, namespace)
}

func (m *multiClusterVeleroClient) GetBackupRepositories(ctx context.Context, namespace string) ([]velerov1.BackupRepository, error) {
	if m.sourceClient == nil {
		return nil, fmt.Errorf("source client not configured")
	}
	return m.sourceClient.GetBackupRepositories(ctx, namespace)
}

func (m *multiClusterVeleroClient) GetBackupStorageLocations(ctx context.Context, namespace string) ([]velerov1.BackupStorageLocation, error) {
	if m.sourceClient == nil {
		return nil, fmt.Errorf("source client not configured")
	}
	return m.sourceClient.GetBackupStorageLocations(ctx, namespace)
}

func (m *multiClusterVeleroClient) GetVolumeSnapshotLocations(ctx context.Context, namespace string) ([]velerov1.VolumeSnapshotLocation, error) {
	if m.sourceClient == nil {
		return nil, fmt.Errorf("source client not configured")
	}
	return m.sourceClient.GetVolumeSnapshotLocations(ctx, namespace)
}

func (m *multiClusterVeleroClient) GetPodVolumeRestores(ctx context.Context, namespace string) ([]velerov1.PodVolumeRestore, error) {
	if m.sourceClient == nil {
		return nil, fmt.Errorf("source client not configured")
	}
	return m.sourceClient.GetPodVolumeRestores(ctx, namespace)
}

func (m *multiClusterVeleroClient) GetDownloadRequests(ctx context.Context, namespace string) ([]velerov1.DownloadRequest, error) {
	if m.sourceClient == nil {
		return nil, fmt.Errorf("source client not configured")
	}
	return m.sourceClient.GetDownloadRequests(ctx, namespace)
}

func (m *multiClusterVeleroClient) GetDataUploads(ctx context.Context, namespace string) ([]velerov2.DataUpload, error) {
	if m.sourceClient == nil {
		return nil, fmt.Errorf("source client not configured")
	}
	return m.sourceClient.GetDataUploads(ctx, namespace)
}

func (m *multiClusterVeleroClient) GetDataDownloads(ctx context.Context, namespace string) ([]velerov2.DataDownload, error) {
	if m.sourceClient == nil {
		return nil, fmt.Errorf("source client not configured")
	}
	return m.sourceClient.GetDataDownloads(ctx, namespace)
}

func (m *multiClusterVeleroClient) GetServerStatusRequests(ctx context.Context, namespace string) ([]velerov1.ServerStatusRequest, error) {
	if m.sourceClient == nil {
		return nil, fmt.Errorf("source client not configured")
	}
	return m.sourceClient.GetServerStatusRequests(ctx, namespace)
}

// src, dst cluster
func (m *multiClusterVeleroClient) ValidateDestinationCluster(ctx context.Context, destConfig string) error {
	if destConfig == "" && m.destinationClient == nil {
		return fmt.Errorf("destination cluster not configured")
	}

	var client interfaces.VeleroService
	var err error

	if destConfig != "" {
		client, err = NewVeleroClientFromRawConfig(destConfig)
		if err != nil {
			return fmt.Errorf("failed to create destination client: %w", err)
		}
	} else {
		client = m.destinationClient
	}

	// Velero가 설치되어 있는지 확인
	_, err = client.GetBackupStorageLocations(ctx, "velero")
	return err
}

func (m *multiClusterVeleroClient) MigrateBackup(ctx context.Context, backupName string, destConfig string) error {
	// 백업 마이그레이션 로직 구현
	// 1. 소스에서 백업 정보 조회
	// 2. 대상 클러스터에서 복원 작업 생성
	// 3. 진행 상황 모니터링

	if m.sourceClient == nil {
		return fmt.Errorf("source client not configured")
	}

	// 구현 예시 (실제로는 더 복잡한 로직 필요)
	backups, err := m.sourceClient.GetBackups(ctx, m.config.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get backups from source: %w", err)
	}

	// 백업 찾기
	var targetBackup *velerov1.Backup
	for _, backup := range backups {
		if backup.Name == backupName {
			targetBackup = &backup
			break
		}
	}

	if targetBackup == nil {
		return fmt.Errorf("backup %s not found", backupName)
	}

	// TODO: 실제 마이그레이션 로직 구현
	return fmt.Errorf("migration not implemented yet")
}

func (m *multiClusterVeleroClient) CompareStorageClasses(ctx context.Context, destConfig string) (*interfaces.StorageClassComparison, error) {
	// 스토리지 클래스 비교 로직 구현
	// TODO: 실제 구현 필요
	return &interfaces.StorageClassComparison{
		Compatible: false,
	}, fmt.Errorf("storage class comparison not implemented yet")
}
