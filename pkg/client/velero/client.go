package velero

import (
	"context"
	"fmt"

	"github.com/taking/kubemigrate/internal/validator"
	"github.com/taking/kubemigrate/pkg/config"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Client Velero 클라이언트 인터페이스
type Client interface {
	// Backup 관련
	GetBackups(ctx context.Context, namespace string) ([]velerov1.Backup, error)
	GetBackup(ctx context.Context, namespace, name string) (*velerov1.Backup, error)
	CreateBackup(ctx context.Context, namespace string, backup *velerov1.Backup) error
	DeleteBackup(ctx context.Context, namespace, name string) error

	// Restore 관련
	GetRestores(ctx context.Context, namespace string) ([]velerov1.Restore, error)
	GetRestore(ctx context.Context, namespace, name string) (*velerov1.Restore, error)
	CreateRestore(ctx context.Context, namespace string, restore *velerov1.Restore) error
	DeleteRestore(ctx context.Context, namespace, name string) error

	// BackupRepository 관련
	GetBackupRepositories(ctx context.Context, namespace string) ([]velerov1.BackupRepository, error)
	GetBackupRepository(ctx context.Context, namespace, name string) (*velerov1.BackupRepository, error)

	// BackupStorageLocation 관련
	GetBackupStorageLocations(ctx context.Context, namespace string) ([]velerov1.BackupStorageLocation, error)
	GetBackupStorageLocation(ctx context.Context, namespace, name string) (*velerov1.BackupStorageLocation, error)
	CreateBackupStorageLocation(ctx context.Context, namespace string, bsl *velerov1.BackupStorageLocation) error
	DeleteBackupStorageLocation(ctx context.Context, namespace, name string) error

	// VolumeSnapshotLocation 관련
	GetVolumeSnapshotLocations(ctx context.Context, namespace string) ([]velerov1.VolumeSnapshotLocation, error)
	GetVolumeSnapshotLocation(ctx context.Context, namespace, name string) (*velerov1.VolumeSnapshotLocation, error)

	// PodVolumeRestore 관련
	GetPodVolumeRestores(ctx context.Context, namespace string) ([]velerov1.PodVolumeRestore, error)
	GetPodVolumeRestore(ctx context.Context, namespace, name string) (*velerov1.PodVolumeRestore, error)

	// HealthCheck : Velero 연결 확인
	HealthCheck(ctx context.Context) error
}

// client Velero 클라이언트 구현체
type client struct {
	k8sClient ctrlclient.Client
	clientset *kubernetes.Clientset
}

// NewClient : 새로운 Velero 클라이언트를 생성합니다 (기본 설정)
func NewClient() (Client, error) {
	// 기본적으로 in-cluster config를 사용
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		// in-cluster config가 없으면 kubeconfig 파일을 사용
		restConfig, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Velero 스키마 등록
	if err := velerov1.AddToScheme(scheme.Scheme); err != nil {
		return nil, fmt.Errorf("failed to add velero scheme: %w", err)
	}

	k8sClient, err := ctrlclient.New(restConfig, ctrlclient.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create controller client: %w", err)
	}

	return &client{
		k8sClient: k8sClient,
		clientset: clientset,
	}, nil
}

// NewClientWithConfig 설정을 받아서 Velero 클라이언트를 생성합니다
func NewClientWithConfig(cfg config.VeleroConfig) (Client, error) {
	var restConfig *rest.Config
	var err error

	if cfg.KubeConfig.KubeConfig != "" {
		// 외부 kubeconfig 사용 (base64 디코딩)
		decodedKubeConfig, err := validator.DecodeIfBase64(cfg.KubeConfig.KubeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to decode kubeconfig: %w", err)
		}

		restConfig, err = clientcmd.RESTConfigFromKubeConfig([]byte(decodedKubeConfig))
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
		}
	} else {
		// 클러스터 내부 설정 사용
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	// Velero 스키마 등록
	if err := velerov1.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	k8sClient, err := ctrlclient.New(restConfig, ctrlclient.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, err
	}

	return &client{
		k8sClient: k8sClient,
		clientset: clientset,
	}, nil
}

// GetBackups 네임스페이스의 Backup 목록을 조회합니다
func (c *client) GetBackups(ctx context.Context, namespace string) ([]velerov1.Backup, error) {
	var backupList velerov1.BackupList
	err := c.k8sClient.List(ctx, &backupList, ctrlclient.InNamespace(namespace))
	if err != nil {
		return nil, err
	}
	return backupList.Items, nil
}

// GetBackup 특정 Backup을 조회합니다
func (c *client) GetBackup(ctx context.Context, namespace, name string) (*velerov1.Backup, error) {
	var backup velerov1.Backup
	err := c.k8sClient.Get(ctx, ctrlclient.ObjectKey{Namespace: namespace, Name: name}, &backup)
	if err != nil {
		return nil, err
	}
	return &backup, nil
}

// CreateBackup Backup을 생성합니다
func (c *client) CreateBackup(ctx context.Context, namespace string, backup *velerov1.Backup) error {
	backup.Namespace = namespace
	return c.k8sClient.Create(ctx, backup)
}

// DeleteBackup Backup을 삭제합니다
func (c *client) DeleteBackup(ctx context.Context, namespace, name string) error {
	backup := &velerov1.Backup{}
	backup.Namespace = namespace
	backup.Name = name
	return c.k8sClient.Delete(ctx, backup)
}

// GetRestores 네임스페이스의 Restore 목록을 조회합니다
func (c *client) GetRestores(ctx context.Context, namespace string) ([]velerov1.Restore, error) {
	var restoreList velerov1.RestoreList
	err := c.k8sClient.List(ctx, &restoreList, ctrlclient.InNamespace(namespace))
	if err != nil {
		return nil, err
	}
	return restoreList.Items, nil
}

// GetRestore 특정 Restore를 조회합니다
func (c *client) GetRestore(ctx context.Context, namespace, name string) (*velerov1.Restore, error) {
	var restore velerov1.Restore
	err := c.k8sClient.Get(ctx, ctrlclient.ObjectKey{Namespace: namespace, Name: name}, &restore)
	if err != nil {
		return nil, err
	}
	return &restore, nil
}

// CreateRestore Restore를 생성합니다
func (c *client) CreateRestore(ctx context.Context, namespace string, restore *velerov1.Restore) error {
	restore.Namespace = namespace
	return c.k8sClient.Create(ctx, restore)
}

// DeleteRestore Restore를 삭제합니다
func (c *client) DeleteRestore(ctx context.Context, namespace, name string) error {
	restore := &velerov1.Restore{}
	restore.Namespace = namespace
	restore.Name = name
	return c.k8sClient.Delete(ctx, restore)
}

// GetBackupRepositories 네임스페이스의 BackupRepository 목록을 조회합니다
func (c *client) GetBackupRepositories(ctx context.Context, namespace string) ([]velerov1.BackupRepository, error) {
	var repoList velerov1.BackupRepositoryList
	err := c.k8sClient.List(ctx, &repoList, ctrlclient.InNamespace(namespace))
	if err != nil {
		return nil, err
	}
	return repoList.Items, nil
}

// GetBackupRepository 특정 BackupRepository를 조회합니다
func (c *client) GetBackupRepository(ctx context.Context, namespace, name string) (*velerov1.BackupRepository, error) {
	var repo velerov1.BackupRepository
	err := c.k8sClient.Get(ctx, ctrlclient.ObjectKey{Namespace: namespace, Name: name}, &repo)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

// GetBackupStorageLocations 네임스페이스의 BackupStorageLocation 목록을 조회합니다
func (c *client) GetBackupStorageLocations(ctx context.Context, namespace string) ([]velerov1.BackupStorageLocation, error) {
	var locationList velerov1.BackupStorageLocationList
	err := c.k8sClient.List(ctx, &locationList, ctrlclient.InNamespace(namespace))
	if err != nil {
		return nil, err
	}
	return locationList.Items, nil
}

// GetBackupStorageLocation 특정 BackupStorageLocation을 조회합니다
func (c *client) GetBackupStorageLocation(ctx context.Context, namespace, name string) (*velerov1.BackupStorageLocation, error) {
	var location velerov1.BackupStorageLocation
	err := c.k8sClient.Get(ctx, ctrlclient.ObjectKey{Namespace: namespace, Name: name}, &location)
	if err != nil {
		return nil, err
	}
	return &location, nil
}

// GetVolumeSnapshotLocations 네임스페이스의 VolumeSnapshotLocation 목록을 조회합니다
func (c *client) GetVolumeSnapshotLocations(ctx context.Context, namespace string) ([]velerov1.VolumeSnapshotLocation, error) {
	var locationList velerov1.VolumeSnapshotLocationList
	err := c.k8sClient.List(ctx, &locationList, ctrlclient.InNamespace(namespace))
	if err != nil {
		return nil, err
	}
	return locationList.Items, nil
}

// GetVolumeSnapshotLocation 특정 VolumeSnapshotLocation을 조회합니다
func (c *client) GetVolumeSnapshotLocation(ctx context.Context, namespace, name string) (*velerov1.VolumeSnapshotLocation, error) {
	var location velerov1.VolumeSnapshotLocation
	err := c.k8sClient.Get(ctx, ctrlclient.ObjectKey{Namespace: namespace, Name: name}, &location)
	if err != nil {
		return nil, err
	}
	return &location, nil
}

// GetPodVolumeRestores 네임스페이스의 PodVolumeRestore 목록을 조회합니다
func (c *client) GetPodVolumeRestores(ctx context.Context, namespace string) ([]velerov1.PodVolumeRestore, error) {
	var restoreList velerov1.PodVolumeRestoreList
	err := c.k8sClient.List(ctx, &restoreList, ctrlclient.InNamespace(namespace))
	if err != nil {
		return nil, err
	}
	return restoreList.Items, nil
}

// GetPodVolumeRestore 특정 PodVolumeRestore를 조회합니다
func (c *client) GetPodVolumeRestore(ctx context.Context, namespace, name string) (*velerov1.PodVolumeRestore, error) {
	var restore velerov1.PodVolumeRestore
	err := c.k8sClient.Get(ctx, ctrlclient.ObjectKey{Namespace: namespace, Name: name}, &restore)
	if err != nil {
		return nil, err
	}
	return &restore, nil
}

// CreateBackupStorageLocation BackupStorageLocation을 생성합니다
func (c *client) CreateBackupStorageLocation(ctx context.Context, namespace string, bsl *velerov1.BackupStorageLocation) error {
	return c.k8sClient.Create(ctx, bsl)
}

// DeleteBackupStorageLocation BackupStorageLocation을 삭제합니다
func (c *client) DeleteBackupStorageLocation(ctx context.Context, namespace, name string) error {
	bsl := &velerov1.BackupStorageLocation{}
	bsl.Name = name
	bsl.Namespace = namespace
	return c.k8sClient.Delete(ctx, bsl)
}

// HealthCheck : Velero 연결 확인
func (c *client) HealthCheck(ctx context.Context) error {
	// 간단한 API 호출로 Velero 연결 상태 확인
	_, err := c.GetBackups(ctx, "velero")
	return err
}
