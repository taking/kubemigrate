package interfaces

import (
	"context"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	velerov2 "github.com/vmware-tanzu/velero/pkg/apis/velero/v2alpha1"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

// HealthChecker : 헬스체크 기능을 제공하는 인터페이스
type HealthChecker interface {
	HealthCheck(ctx context.Context) error // 리소스 또는 클러스터 연결 상태 확인
}

// KubernetesClient : Kubernetes 클러스터 관련 작업을 정의하는 인터페이스
type KubernetesClient interface {
	HealthChecker
	GetPods(ctx context.Context) ([]v1.Pod, error)                           // 지정된 네임스페이스에서 Pod 목록 조회
	GetStorageClasses(ctx context.Context) ([]storagev1.StorageClass, error) // 스토리지 클래스 목록 조회
}

// VeleroClient : Velero 관련 작업을 정의하는 인터페이스
type VeleroClient interface {
	HealthChecker
	GetBackups(ctx context.Context) ([]velerov1.Backup, error)                                 // Velero Backup 목록 조회
	GetRestores(ctx context.Context) ([]velerov1.Restore, error)                               // Velero Restore 목록 조회
	GetBackupRepositories(ctx context.Context) ([]velerov1.BackupRepository, error)            // BackupRepository 목록 조회
	GetBackupStorageLocations(ctx context.Context) ([]velerov1.BackupStorageLocation, error)   // BackupStorageLocation 목록 조회
	GetVolumeSnapshotLocations(ctx context.Context) ([]velerov1.VolumeSnapshotLocation, error) // VolumeSnapshotLocation 목록 조회
	GetPodVolumeRestores(ctx context.Context) ([]velerov1.PodVolumeRestore, error)             // PodVolumeRestore 목록 조회
	GetDownloadRequests(ctx context.Context) ([]velerov1.DownloadRequest, error)               // DownloadRequest 목록 조회
	GetDataUploads(ctx context.Context) ([]velerov2.DataUpload, error)                         // DataUpload 목록 조회
	GetDataDownloads(ctx context.Context) ([]velerov2.DataDownload, error)                     // DataDownload 목록 조회
	GetDeleteBackupRequests(ctx context.Context) ([]velerov1.DeleteBackupRequest, error)       // DeleteBackupRequest 목록 조회
	GetServerStatusRequests(ctx context.Context) ([]velerov1.ServerStatusRequest, error)       // ServerStatusRequest 목록 조회
}

// HelmClient : Helm 관련 작업을 정의하는 인터페이스
type HelmClient interface {
	HealthChecker
	IsChartInstalled(chartName string) (bool, *release.Release, error)             // 차트가 설치되어 있는지 확인
	InstallChart(chartName, chartPath string, values map[string]interface{}) error // Helm 차트 설치
	InvalidateCache() error                                                        // Helm 캐시 무효화
}

// MinioClient : MinIO 관련 작업을 정의하는 인터페이스
type MinioClient interface {
	HealthChecker
	BucketExists(ctx context.Context, bucketName string) (bool, error)                      // 버킷 존재 확인
	CreateBucket(ctx context.Context, bucketName string) error                              // 버킷 생성
	CreateBucketIfNotExists(ctx context.Context, bucketName, region string) (string, error) // 버킷이 존재하지 않으면 생성
}
