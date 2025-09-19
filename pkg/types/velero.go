package types

import (
	"fmt"

	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

// Velero 리소스 타입 정의
type (
	// Backup 관련
	BackupList     = []velerov1.Backup
	Backup         = velerov1.Backup
	BackupResource interface {
		BackupList | *Backup
	}

	// Restore 관련
	RestoreList     = []velerov1.Restore
	Restore         = velerov1.Restore
	RestoreResource interface {
		RestoreList | *Restore
	}

	// BackupStorageLocation 관련
	BackupStorageLocationList     = []velerov1.BackupStorageLocation
	BackupStorageLocation         = velerov1.BackupStorageLocation
	BackupStorageLocationResource interface {
		BackupStorageLocationList | *BackupStorageLocation
	}

	// VolumeSnapshotLocation 관련
	VolumeSnapshotLocationList     = []velerov1.VolumeSnapshotLocation
	VolumeSnapshotLocation         = velerov1.VolumeSnapshotLocation
	VolumeSnapshotLocationResource interface {
		VolumeSnapshotLocationList | *VolumeSnapshotLocation
	}

	// BackupRepository 관련
	BackupRepositoryList     = []velerov1.BackupRepository
	BackupRepository         = velerov1.BackupRepository
	BackupRepositoryResource interface {
		BackupRepositoryList | *BackupRepository
	}

	// PodVolumeRestore 관련
	PodVolumeRestoreList     = []velerov1.PodVolumeRestore
	PodVolumeRestore         = velerov1.PodVolumeRestore
	PodVolumeRestoreResource interface {
		PodVolumeRestoreList | *PodVolumeRestore
	}
)

// Velero 타입 어설션 헬퍼 함수들

// 타입 어설션 헬퍼 함수들
func AssertBackupList(v interface{}) (BackupList, bool) {
	backupList, ok := v.(BackupList)
	return backupList, ok
}

func AssertBackup(v interface{}) (*Backup, bool) {
	backup, ok := v.(*Backup)
	return backup, ok
}

func AssertRestoreList(v interface{}) (RestoreList, bool) {
	restoreList, ok := v.(RestoreList)
	return restoreList, ok
}

func AssertRestore(v interface{}) (*Restore, bool) {
	restore, ok := v.(*Restore)
	return restore, ok
}

func AssertBackupStorageLocationList(v interface{}) (BackupStorageLocationList, bool) {
	bslList, ok := v.(BackupStorageLocationList)
	return bslList, ok
}

func AssertBackupStorageLocation(v interface{}) (*BackupStorageLocation, bool) {
	bsl, ok := v.(*BackupStorageLocation)
	return bsl, ok
}

func AssertVolumeSnapshotLocationList(v interface{}) (VolumeSnapshotLocationList, bool) {
	vslList, ok := v.(VolumeSnapshotLocationList)
	return vslList, ok
}

func AssertVolumeSnapshotLocation(v interface{}) (*VolumeSnapshotLocation, bool) {
	vsl, ok := v.(*VolumeSnapshotLocation)
	return vsl, ok
}

func AssertBackupRepositoryList(v interface{}) (BackupRepositoryList, bool) {
	repoList, ok := v.(BackupRepositoryList)
	return repoList, ok
}

func AssertBackupRepository(v interface{}) (*BackupRepository, bool) {
	repo, ok := v.(*BackupRepository)
	return repo, ok
}

func AssertPodVolumeRestoreList(v interface{}) (PodVolumeRestoreList, bool) {
	pvrList, ok := v.(PodVolumeRestoreList)
	return pvrList, ok
}

func AssertPodVolumeRestore(v interface{}) (*PodVolumeRestore, bool) {
	pvr, ok := v.(*PodVolumeRestore)
	return pvr, ok
}

// 안전한 타입 어설션을 위한 래퍼 함수들
func SafeGetBackupList(v interface{}) (BackupList, error) {
	if backupList, ok := AssertBackupList(v); ok {
		return backupList, nil
	}
	return nil, fmt.Errorf("expected BackupList, got %T", v)
}

func SafeGetBackup(v interface{}) (*Backup, error) {
	if backup, ok := AssertBackup(v); ok {
		return backup, nil
	}
	return nil, fmt.Errorf("expected *Backup, got %T", v)
}

func SafeGetRestoreList(v interface{}) (RestoreList, error) {
	if restoreList, ok := AssertRestoreList(v); ok {
		return restoreList, nil
	}
	return nil, fmt.Errorf("expected RestoreList, got %T", v)
}

func SafeGetRestore(v interface{}) (*Restore, error) {
	if restore, ok := AssertRestore(v); ok {
		return restore, nil
	}
	return nil, fmt.Errorf("expected *Restore, got %T", v)
}

func SafeGetBackupStorageLocationList(v interface{}) (BackupStorageLocationList, error) {
	if bslList, ok := AssertBackupStorageLocationList(v); ok {
		return bslList, nil
	}
	return nil, fmt.Errorf("expected BackupStorageLocationList, got %T", v)
}

func SafeGetBackupStorageLocation(v interface{}) (*BackupStorageLocation, error) {
	if bsl, ok := AssertBackupStorageLocation(v); ok {
		return bsl, nil
	}
	return nil, fmt.Errorf("expected *BackupStorageLocation, got %T", v)
}

func SafeGetVolumeSnapshotLocationList(v interface{}) (VolumeSnapshotLocationList, error) {
	if vslList, ok := AssertVolumeSnapshotLocationList(v); ok {
		return vslList, nil
	}
	return nil, fmt.Errorf("expected VolumeSnapshotLocationList, got %T", v)
}

func SafeGetVolumeSnapshotLocation(v interface{}) (*VolumeSnapshotLocation, error) {
	if vsl, ok := AssertVolumeSnapshotLocation(v); ok {
		return vsl, nil
	}
	return nil, fmt.Errorf("expected *VolumeSnapshotLocation, got %T", v)
}

func SafeGetBackupRepositoryList(v interface{}) (BackupRepositoryList, error) {
	if repoList, ok := AssertBackupRepositoryList(v); ok {
		return repoList, nil
	}
	return nil, fmt.Errorf("expected BackupRepositoryList, got %T", v)
}

func SafeGetBackupRepository(v interface{}) (*BackupRepository, error) {
	if repo, ok := AssertBackupRepository(v); ok {
		return repo, nil
	}
	return nil, fmt.Errorf("expected *BackupRepository, got %T", v)
}

func SafeGetPodVolumeRestoreList(v interface{}) (PodVolumeRestoreList, error) {
	if pvrList, ok := AssertPodVolumeRestoreList(v); ok {
		return pvrList, nil
	}
	return nil, fmt.Errorf("expected PodVolumeRestoreList, got %T", v)
}

func SafeGetPodVolumeRestore(v interface{}) (*PodVolumeRestore, error) {
	if pvr, ok := AssertPodVolumeRestore(v); ok {
		return pvr, nil
	}
	return nil, fmt.Errorf("expected *PodVolumeRestore, got %T", v)
}
