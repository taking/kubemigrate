package interfaces

//import (
//	"context"
//	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
//	velerov2 "github.com/vmware-tanzu/velero/pkg/apis/velero/v2alpha1"
//)
//
//type VeleroService interface {
//	// Velero 백업 객체 목록을 조회합니다.
//	GetBackups(ctx context.Context) ([]velerov1.Backup, error)
//
//	// Velero 복원(Restore) 객체 목록을 조회합니다.
//	GetRestores(ctx context.Context) ([]velerov1.Restore, error)
//
//	// 백업 저장소(Repository) 상태 객체 목록을 조회합니다.
//	GetBackupRepositories(ctx context.Context) ([]velerov1.BackupRepository, error)
//
//	// 백업 스토리지 위치(BackupStorageLocation) 목록을 조회합니다.
//	GetBackupStorageLocations(ctx context.Context) ([]velerov1.BackupStorageLocation, error)
//
//	// 볼륨 스냅샷 저장소 위치 목록을 조회합니다.
//	GetVolumeSnapshotLocations(ctx context.Context) ([]velerov1.VolumeSnapshotLocation, error)
//
//	// 파드 볼륨 복원(PodVolumeRestore) 객체 목록을 조회합니다.
//	GetPodVolumeRestores(ctx context.Context) ([]velerov1.PodVolumeRestore, error)
//
//	// 백업 파일 다운로드 요청 객체 목록을 조회합니다.
//	GetDownloadRequests(ctx context.Context) ([]velerov1.DownloadRequest, error)
//
//	// DataUpload(v2) 객체 목록을 조회하여 백업 업로드 상태를 확인합니다.
//	GetDataUploads(ctx context.Context) ([]velerov2.DataUpload, error)
//
//	// DataDownload(v2) 객체 목록을 조회하여 백업 다운로드 상태를 확인합니다.
//	GetDataDownloads(ctx context.Context) ([]velerov2.DataDownload, error)
//
//	// 서버 상태 확인 요청(ServerStatusRequest) 목록을 조회합니다.
//	GetServerStatusRequests(ctx context.Context) ([]velerov1.ServerStatusRequest, error)
//}
