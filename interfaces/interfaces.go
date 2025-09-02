package interfaces

import (
	"context"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	velerov2 "github.com/vmware-tanzu/velero/pkg/apis/velero/v2alpha1"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

// HealthChecker defines health check capability
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// KubernetesClient defines Kubernetes operations
type KubernetesClient interface {
	HealthChecker
	GetPods(ctx context.Context) ([]v1.Pod, error)
	GetStorageClasses(ctx context.Context) ([]storagev1.StorageClass, error)
}

// VeleroClient defines Velero operations
type VeleroClient interface {
	HealthChecker
	GetBackups(ctx context.Context) ([]velerov1.Backup, error)
	GetRestores(ctx context.Context) ([]velerov1.Restore, error)
	GetBackupRepositories(ctx context.Context) ([]velerov1.BackupRepository, error)
	GetBackupStorageLocations(ctx context.Context) ([]velerov1.BackupStorageLocation, error)
	GetVolumeSnapshotLocations(ctx context.Context) ([]velerov1.VolumeSnapshotLocation, error)
	GetPodVolumeRestores(ctx context.Context) ([]velerov1.PodVolumeRestore, error)
	GetDownloadRequests(ctx context.Context) ([]velerov1.DownloadRequest, error)
	GetDataUploads(ctx context.Context) ([]velerov2.DataUpload, error)
	GetDataDownloads(ctx context.Context) ([]velerov2.DataDownload, error)
	GetDeleteBackupRequests(ctx context.Context) ([]velerov1.DeleteBackupRequest, error)
	GetServerStatusRequests(ctx context.Context) ([]velerov1.ServerStatusRequest, error)
}

// HelmClient defines Helm operations
type HelmClient interface {
	HealthChecker
	IsChartInstalled(chartName string) (bool, *release.Release, error)
	InstallChart(chartName, chartPath string, values map[string]interface{}) error
	InvalidateCache() error
}

// MinioClient defines MinIO operations
type MinioClient interface {
	HealthChecker
	CreateBucketIfNotExists(ctx context.Context, bucketName, region string) (string, error)
}
