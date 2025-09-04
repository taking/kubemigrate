package client

import (
	"context"
	"fmt"
	"time"

	"github.com/taking/kubemigrate/pkg/interfaces"
	"github.com/taking/kubemigrate/pkg/models"
	"github.com/taking/kubemigrate/pkg/utils"

	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	velerov2 "github.com/vmware-tanzu/velero/pkg/apis/velero/v2alpha1"
	"github.com/vmware-tanzu/velero/pkg/client"
	kbclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/taking/kubemigrate/pkg/errors"
)

// veleroClient : Velero 클라이언트
type veleroClient struct {
	client  kbclient.Client
	ns      string
	factory *ClientFactory
}

// NewVeleroClient : Velero 클라이언트 초기화
func NewVeleroClient(cfg models.KubeConfig) (interfaces.VeleroClient, error) {
	clientFactory := NewClientFactory()

	restCfg, err := clientFactory.CreateRESTConfig(cfg)
	if err != nil {
		return nil, err
	}

	ns := getNamespaceOrDefault(cfg.Namespace, "velero")

	vcl := client.VeleroConfig{
		"KubeClientConfig": restCfg,
		"Namespace":        ns,
	}

	veleroFactory := client.NewFactory("velero", vcl)
	kb, err := veleroFactory.KubebuilderClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create velero factory kubebuilder client: %w", err)
	}

	return &veleroClient{
		client:  kb,
		ns:      ns,
		factory: clientFactory,
	}, nil
}

// HealthCheck : Kubernetes 연결 확인
func (v *veleroClient) HealthCheck(ctx context.Context) error {
	// 5초 제한 타임아웃
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 서버 연결 확인 (백업 목록 조회 시도)
	var backups velerov1.BackupList
	return utils.RunWithTimeout(ctx, func() error {
		return v.client.List(ctx, &backups)
	})
}

// GetBackups : 백업 목록 조회
func (v *veleroClient) GetBackups(ctx context.Context) ([]velerov1.Backup, error) {
	list := &velerov1.BackupList{}

	// 목록 조회
	if err := v.client.List(ctx, list, kbclient.InNamespace(v.ns)); err != nil {
		return nil, errors.WrapK8sError(v.ns, err, "backups")
	}

	// ManagedFields 제외
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}
	return list.Items, nil
}

func (v *veleroClient) GetRestores(ctx context.Context) ([]velerov1.Restore, error) {
	list := &velerov1.RestoreList{}

	// 목록 조회
	if err := v.client.List(ctx, list, kbclient.InNamespace(v.ns)); err != nil {
		return nil, errors.WrapK8sError(v.ns, err, "restores")
	}

	// ManagedFields 제외
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}
	return list.Items, nil
}

func (v *veleroClient) GetBackupRepositories(ctx context.Context) ([]velerov1.BackupRepository, error) {
	list := &velerov1.BackupRepositoryList{}

	// 목록 조회
	if err := v.client.List(ctx, list, kbclient.InNamespace(v.ns)); err != nil {
		return nil, errors.WrapK8sError(v.ns, err, "backupRepositories")
	}

	// ManagedFields 제외
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}
	return list.Items, nil
}

func (v *veleroClient) GetBackupStorageLocations(ctx context.Context) ([]velerov1.BackupStorageLocation, error) {
	list := &velerov1.BackupStorageLocationList{}

	// 목록 조회
	if err := v.client.List(ctx, list, kbclient.InNamespace(v.ns)); err != nil {
		return nil, errors.WrapK8sError(v.ns, err, "backupStorageLocations")
	}

	// ManagedFields 제외
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}
	return list.Items, nil
}

func (v *veleroClient) GetVolumeSnapshotLocations(ctx context.Context) ([]velerov1.VolumeSnapshotLocation, error) {
	list := &velerov1.VolumeSnapshotLocationList{}

	// 목록 조회
	if err := v.client.List(ctx, list, kbclient.InNamespace(v.ns)); err != nil {
		return nil, errors.WrapK8sError(v.ns, err, "volumeSnapshotLocations")
	}

	// ManagedFields 제외
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}
	return list.Items, nil
}

func (v *veleroClient) GetPodVolumeRestores(ctx context.Context) ([]velerov1.PodVolumeRestore, error) {
	list := &velerov1.PodVolumeRestoreList{}

	// 목록 조회
	if err := v.client.List(ctx, list, kbclient.InNamespace(v.ns)); err != nil {
		return nil, errors.WrapK8sError(v.ns, err, "podVolumeRestores")
	}

	// ManagedFields 제외
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}
	return list.Items, nil
}

func (v *veleroClient) GetDownloadRequests(ctx context.Context) ([]velerov1.DownloadRequest, error) {
	list := &velerov1.DownloadRequestList{}

	// 목록 조회
	if err := v.client.List(ctx, list, kbclient.InNamespace(v.ns)); err != nil {
		return nil, errors.WrapK8sError(v.ns, err, "downloadRequests")
	}

	// ManagedFields 제외
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}
	return list.Items, nil
}

func (v *veleroClient) GetDataUploads(ctx context.Context) ([]velerov2.DataUpload, error) {
	list := &velerov2.DataUploadList{}

	// 목록 조회
	if err := v.client.List(ctx, list, kbclient.InNamespace(v.ns)); err != nil {
		return nil, errors.WrapK8sError(v.ns, err, "dataUploads")
	}

	// ManagedFields 제외
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}
	return list.Items, nil
}

func (v *veleroClient) GetDataDownloads(ctx context.Context) ([]velerov2.DataDownload, error) {
	list := &velerov2.DataDownloadList{}

	// 목록 조회
	if err := v.client.List(ctx, list, kbclient.InNamespace(v.ns)); err != nil {
		return nil, errors.WrapK8sError(v.ns, err, "dataDownloads")
	}

	// ManagedFields 제외
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}
	return list.Items, nil
}

func (v *veleroClient) GetDeleteBackupRequests(ctx context.Context) ([]velerov1.DeleteBackupRequest, error) {
	list := &velerov1.DeleteBackupRequestList{}

	// 목록 조회
	if err := v.client.List(ctx, list, kbclient.InNamespace(v.ns)); err != nil {
		return nil, errors.WrapK8sError(v.ns, err, "deleteBackupRequests")
	}

	// ManagedFields 제외
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}
	return list.Items, nil
}

func (v *veleroClient) GetServerStatusRequests(ctx context.Context) ([]velerov1.ServerStatusRequest, error) {
	list := &velerov1.ServerStatusRequestList{}

	// 목록 조회
	if err := v.client.List(ctx, list, kbclient.InNamespace(v.ns)); err != nil {
		return nil, errors.WrapK8sError(v.ns, err, "serverStatusRequests")
	}

	// ManagedFields 제외
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}
	return list.Items, nil
}

//
//// CreateBackupStorageLocation ensures a BackupStorageLocation exists.
//// If it exists, returns a message. If not, creates it. Returns error if creation fails.
//func (v *veleroClient) EnsureBackupStorageLocation(ctx context.Context, namespace, name, provider, bucket, prefix, region string) (*velerov1.BackupStorageLocation, string, error) {
//	c, err := v.newKBClient()
//	if err != nil {
//		return nil, "", fmt.Errorf("failed to get Kubernetes client: %w", err)
//	}
//
//	// 1. 기존 BSL 조회
//	existing := &velerov1.BackupStorageLocation{}
//	err = c.Get(ctx, kbclient.ObjectKey{Namespace: namespace, Name: name}, existing)
//	if err == nil {
//		// 이미 존재함
//		utils.StripManagedFields(existing)
//		return existing, fmt.Sprintf("BackupStorageLocation %q already exists", name), nil
//	}
//
//	// 존재하지 않으면 새로 생성
//	newBSL := &velerov1.BackupStorageLocation{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      name,
//			Namespace: namespace,
//		},
//		Spec: velerov1.BackupStorageLocationSpec{
//			Provider: provider,
//			Config: map[string]string{
//				"bucket":           "default",
//				"s3ForcePathStyle": "true",
//				"s3Url":            "https://velero.io",
//			},
//			ObjectStorage: velerov1.ObjectStorageLocation{
//				Bucket: bucket,
//				Prefix: prefix,
//			},
//			StorageType: velerov1.StorageType{
//				ObjectStorage: &velerov1.ObjectStorageLocation{},
//			},
//			// region 옵션은 일부 provider(AWS, S3)에서만 사용
//			// 일부 Provider는 region을 Spec.Config로 전달해야 할 수도 있음
//		},
//	}
//
//	if region != "" {
//		if newBSL.Spec.Config == nil {
//			newBSL.Spec.Config = map[string]string{}
//		}
//		newBSL.Spec.Config["region"] = region
//	}
//
//	if err := c.Create(ctx, newBSL); err != nil {
//		return nil, "", fmt.Errorf("failed to create BackupStorageLocation %q: %w", name, err)
//	}
//
//	pkg.StripManagedFields(newBSL)
//	return newBSL, fmt.Sprintf("BackupStorageLocation %q created successfully", name), nil
//}
