package types

import (
	"fmt"
	"time"

	"github.com/taking/kubemigrate/pkg/config"
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

// Velero 설치 관련 타입들
type (
	// InstallResult : 설치 결과
	InstallResult struct {
		Status           string                 `json:"status"`
		Message          string                 `json:"message"`
		VeleroNamespace  string                 `json:"velero_namespace"`
		MinioConnected   bool                   `json:"minio_connected"`
		BackupLocation   string                 `json:"backup_location"`
		InstallationTime time.Duration          `json:"installation_time"`
		Force            bool                   `json:"force"`
		Details          map[string]interface{} `json:"details,omitempty"`
	}

	// VeleroStatus : Velero 상태 정보
	VeleroStatus struct {
		PodsInstalled    bool   `json:"pods_installed"`
		HelmRelease      bool   `json:"helm_release"`
		ReleaseNamespace string `json:"release_namespace"`
		IsHealthy        bool   `json:"is_healthy"`
		ErrorMessage     string `json:"error_message,omitempty"`
	}

	// InstallationError : 설치 에러 정보
	InstallationError struct {
		Type        string   `json:"type"` // "helm_conflict", "pod_failed", "timeout"
		Message     string   `json:"message"`
		Details     string   `json:"details"`
		Suggestions []string `json:"suggestions"`
		Commands    []string `json:"commands"`
	}

	// InstallationProgress : 설치 진행 상황
	InstallationProgress struct {
		Step      string `json:"step"`
		Progress  int    `json:"progress"` // 0-100
		Message   string `json:"message"`
		Estimated string `json:"estimated"` // 남은 시간
		CanCancel bool   `json:"can_cancel"`
	}

	// BSLStatusInfo : BSL 상태 정보
	BSLStatusInfo struct {
		Phase              string   `json:"phase"`
		Message            string   `json:"message"`
		LastValidationTime string   `json:"last_validation_time"`
		AccessMode         string   `json:"access_mode"`
		StorageType        string   `json:"storage_type"`
		ObjectStorage      string   `json:"object_storage"`
		Credential         string   `json:"credential"`
		Config             string   `json:"config"`
		ValidationErrors   []string `json:"validation_errors,omitempty"`
	}

	// VeleroResources : Velero 리소스 정보
	VeleroResources struct {
		Pods                []string `json:"pods"`
		ConfigMaps          []string `json:"configmaps"`
		Secrets             []string `json:"secrets"`
		Deployments         []string `json:"deployments"`
		DaemonSets          []string `json:"daemonsets"`
		StatefulSets        []string `json:"statefulsets"`
		Services            []string `json:"services"`
		ServiceAccounts     []string `json:"serviceaccounts"`
		ClusterRoles        []string `json:"clusterroles"`
		ClusterRoleBindings []string `json:"clusterrolebindings"`
		CRDs                []string `json:"crds"`
	}

	// BackupRequest : 백업 생성 요청 구조체
	BackupRequest struct {
		Name                     string            `json:"name" binding:"required" example:"my-backup-2024-01-15"`
		Namespace                string            `json:"namespace,omitempty" example:"default"`
		IncludeNamespaces        []string          `json:"includeNamespaces,omitempty" example:"default,kube-system"`
		ExcludeNamespaces        []string          `json:"excludeNamespaces,omitempty" example:"kube-public"`
		IncludeResources         []string          `json:"includeResources,omitempty" example:"pods,services,configmaps"`
		ExcludeResources         []string          `json:"excludeResources,omitempty" example:"events"`
		LabelSelector            map[string]string `json:"labelSelector,omitempty" example:"app=myapp"`
		StorageLocation          string            `json:"storageLocation,omitempty" example:"default"`
		VolumeSnapshotLocations  []string          `json:"volumeSnapshotLocations,omitempty" example:"default"`
		TTL                      string            `json:"ttl,omitempty" example:"720h0m0s"`
		IncludeClusterResources  *bool             `json:"includeClusterResources,omitempty" example:"true"`
		DefaultVolumesToFsBackup *bool             `json:"defaultVolumesToFsBackup,omitempty" example:"false"`
		Hooks                    *BackupHooks      `json:"hooks,omitempty"`
		Metadata                 map[string]string `json:"metadata,omitempty"`
	}

	// CreateBackupRequest : 백업 생성 전체 요청 구조체 (kubeconfig, minio 포함)
	CreateBackupRequest struct {
		KubeConfig  config.KubeConfig  `json:"kubeconfig" binding:"required"`
		MinioConfig config.MinioConfig `json:"minio" binding:"required"`
		Backup      BackupRequest      `json:"backup" binding:"required"`
	}

	// DeleteBackupRequest : 백업 삭제 전체 요청 구조체 (kubeconfig, minio 포함)
	DeleteBackupRequest struct {
		KubeConfig  config.KubeConfig  `json:"kubeconfig" binding:"required"`
		MinioConfig config.MinioConfig `json:"minio" binding:"required"`
	}

	// DeleteRestoreRequest : 복원 삭제 전체 요청 구조체 (kubeconfig, minio 포함)
	DeleteRestoreRequest struct {
		KubeConfig  config.KubeConfig  `json:"kubeconfig" binding:"required"`
		MinioConfig config.MinioConfig `json:"minio" binding:"required"`
	}

	// RestoreRequest : 복원 생성 요청 구조체
	RestoreRequest struct {
		Name                    string            `json:"name" binding:"required" example:"my-restore-2024-01-15"`
		BackupName              string            `json:"backupName" binding:"required" example:"my-backup-2024-01-15"`
		Namespace               string            `json:"namespace,omitempty" example:"default"`
		IncludeNamespaces       []string          `json:"includeNamespaces" binding:"required" example:"default,kube-system"`
		ExcludeNamespaces       []string          `json:"excludeNamespaces,omitempty" example:"kube-public"`
		IncludeResources        []string          `json:"includeResources,omitempty" example:"pods,services,configmaps"`
		ExcludeResources        []string          `json:"excludeResources,omitempty" example:"events"`
		LabelSelector           map[string]string `json:"labelSelector,omitempty" example:"app=myapp"`
		StorageLocation         string            `json:"storageLocation,omitempty" example:"default"`
		VolumeSnapshotLocations []string          `json:"volumeSnapshotLocations,omitempty" example:"default"`
		IncludeClusterResources bool              `json:"includeClusterResources" binding:"required" example:"true"`
		RestorePVs              bool              `json:"restorePVs" binding:"required" example:"true"`
		StorageClassMappings    map[string]string `json:"storageClassMappings,omitempty" example:"original-sc:new-sc"`
		NamespaceMappings       map[string]string `json:"namespaceMappings,omitempty" example:"old-ns:new-ns"`
		Hooks                   *RestoreHooks     `json:"hooks,omitempty"`
		Metadata                map[string]string `json:"metadata,omitempty"`
	}

	// VeleroRestoreRequest : 복원 생성 전체 요청 구조체 (kubeconfig 포함)
	VeleroRestoreRequest struct {
		KubeConfig config.KubeConfig `json:"kubeconfig" binding:"required"`
		Restore    RestoreRequest    `json:"restore" binding:"required"`
	}

	// RestoreHooks : 복원 훅 설정
	RestoreHooks struct {
		Resources []RestoreResourceHookSpec `json:"resources,omitempty"`
	}

	// RestoreResourceHookSpec : 리소스별 복원 훅 설정
	RestoreResourceHookSpec struct {
		Name          string            `json:"name" binding:"required"`
		Namespaces    []string          `json:"namespaces,omitempty"`
		Resources     []string          `json:"resources,omitempty"`
		LabelSelector map[string]string `json:"labelSelector,omitempty"`
		PreHooks      []RestoreHookSpec `json:"preHooks,omitempty"`
		PostHooks     []RestoreHookSpec `json:"postHooks,omitempty"`
	}

	// RestoreHookSpec : 개별 복원 훅 설정
	RestoreHookSpec struct {
		Exec *ExecHook `json:"exec,omitempty"`
	}

	// RestoreResult : 복원 생성 결과
	RestoreResult struct {
		Status      string                 `json:"status"`
		JobID       string                 `json:"jobId"`
		RestoreName string                 `json:"restoreName"`
		BackupName  string                 `json:"backupName"`
		Namespace   string                 `json:"namespace"`
		Message     string                 `json:"message"`
		StatusUrl   string                 `json:"statusUrl"`
		LogsUrl     string                 `json:"logsUrl"`
		CreatedAt   time.Time              `json:"createdAt"`
		Details     map[string]interface{} `json:"details,omitempty"`
	}

	// RestoreValidationResult : 복원 검증 결과
	RestoreValidationResult struct {
		IsValid           bool                     `json:"isValid"`
		RestoreName       string                   `json:"restoreName"`
		BackupName        string                   `json:"backupName"`
		Phase             string                   `json:"phase"`
		ValidationTime    time.Time                `json:"validationTime"`
		ValidationDetails RestoreValidationDetails `json:"validationDetails"`
		Errors            []string                 `json:"errors,omitempty"`
		Warnings          []string                 `json:"warnings,omitempty"`
		Summary           RestoreSummary           `json:"summary"`
	}

	// RestoreValidationDetails : 복원 검증 상세 정보
	RestoreValidationDetails struct {
		BackupExists          bool     `json:"backupExists"`
		BackupValid           bool     `json:"backupValid"`
		StorageLocationValid  bool     `json:"storageLocationValid"`
		VolumeSnapshotValid   bool     `json:"volumeSnapshotValid"`
		ResourceCount         int      `json:"resourceCount"`
		VolumeCount           int      `json:"volumeCount"`
		ValidationErrors      []string `json:"validationErrors,omitempty"`
		BackupErrors          []string `json:"backupErrors,omitempty"`
		StorageLocationErrors []string `json:"storageLocationErrors,omitempty"`
		VolumeSnapshotErrors  []string `json:"volumeSnapshotErrors,omitempty"`
	}

	// RestoreSummary : 복원 요약 정보
	RestoreSummary struct {
		TotalItems       int            `json:"totalItems"`
		TotalSize        string         `json:"totalSize"`
		Duration         string         `json:"duration"`
		StartTime        time.Time      `json:"startTime"`
		EndTime          time.Time      `json:"endTime"`
		ResourceCounts   map[string]int `json:"resourceCounts"`
		VolumeSnapshots  []string       `json:"volumeSnapshots,omitempty"`
		StorageLocation  string         `json:"storageLocation"`
		BackupRepository string         `json:"backupRepository"`
	}

	// BackupHooks : 백업 훅 설정
	BackupHooks struct {
		Resources []BackupResourceHookSpec `json:"resources,omitempty"`
	}

	// BackupResourceHookSpec : 리소스별 훅 설정
	BackupResourceHookSpec struct {
		Name          string            `json:"name" binding:"required"`
		Namespaces    []string          `json:"namespaces,omitempty"`
		Resources     []string          `json:"resources,omitempty"`
		LabelSelector map[string]string `json:"labelSelector,omitempty"`
		PreHooks      []BackupHookSpec  `json:"preHooks,omitempty"`
		PostHooks     []BackupHookSpec  `json:"postHooks,omitempty"`
	}

	// BackupHookSpec : 개별 훅 설정
	BackupHookSpec struct {
		Exec *ExecHook `json:"exec,omitempty"`
	}

	// ExecHook : 실행 훅 설정
	ExecHook struct {
		Container string   `json:"container,omitempty"`
		Command   []string `json:"command" binding:"required"`
		OnError   string   `json:"onError,omitempty"` // "Continue", "Fail"
		Timeout   string   `json:"timeout,omitempty" example:"30s"`
	}

	// BackupResult : 백업 생성 결과
	BackupResult struct {
		Status     string                 `json:"status"`
		JobID      string                 `json:"jobId"`
		BackupName string                 `json:"backupName"`
		Namespace  string                 `json:"namespace"`
		Message    string                 `json:"message"`
		StatusUrl  string                 `json:"statusUrl"`
		LogsUrl    string                 `json:"logsUrl"`
		CreatedAt  time.Time              `json:"createdAt"`
		Details    map[string]interface{} `json:"details,omitempty"`
	}

	// BackupValidationResult : 백업 검증 결과
	BackupValidationResult struct {
		IsValid           bool              `json:"isValid"`
		BackupName        string            `json:"backupName"`
		Phase             string            `json:"phase"`
		ValidationTime    time.Time         `json:"validationTime"`
		ValidationDetails ValidationDetails `json:"validationDetails"`
		Errors            []string          `json:"errors,omitempty"`
		Warnings          []string          `json:"warnings,omitempty"`
		Summary           BackupSummary     `json:"summary"`
	}

	// ValidationDetails : 검증 상세 정보
	ValidationDetails struct {
		StorageLocationValid  bool     `json:"storageLocationValid"`
		VolumeSnapshotValid   bool     `json:"volumeSnapshotValid"`
		BackupRepositoryValid bool     `json:"backupRepositoryValid"`
		ResourceCount         int      `json:"resourceCount"`
		VolumeCount           int      `json:"volumeCount"`
		ValidationErrors      []string `json:"validationErrors,omitempty"`
		StorageLocationErrors []string `json:"storageLocationErrors,omitempty"`
		VolumeSnapshotErrors  []string `json:"volumeSnapshotErrors,omitempty"`
	}

	// BackupSummary : 백업 요약 정보
	BackupSummary struct {
		TotalItems       int            `json:"totalItems"`
		TotalSize        string         `json:"totalSize"`
		Duration         string         `json:"duration"`
		StartTime        time.Time      `json:"startTime"`
		EndTime          time.Time      `json:"endTime"`
		ResourceCounts   map[string]int `json:"resourceCounts"`
		VolumeSnapshots  []string       `json:"volumeSnapshots,omitempty"`
		StorageLocation  string         `json:"storageLocation"`
		BackupRepository string         `json:"backupRepository"`
	}
)

// InstallationError의 Error 인터페이스 구현
func (e *InstallationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Message, e.Details)
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
