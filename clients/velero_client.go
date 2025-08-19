package clients

import (
	"context"
	"fmt"
	"taking.kr/velero/interfaces"

	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	velerov2 "github.com/vmware-tanzu/velero/pkg/apis/velero/v2alpha1"
	"github.com/vmware-tanzu/velero/pkg/client"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kbclient "sigs.k8s.io/controller-runtime/pkg/client"

	"taking.kr/velero/utils"
)

type veleroClient struct {
	factory client.Factory
}

func NewVeleroClientFromRestConfig(config *rest.Config) (interfaces.VeleroService, error) {
	cl := client.VeleroConfig{
		"KubeClientConfig": config,
		"Namespace":        "velero",
	}

	factory := client.NewFactory("velero", cl)
	return &veleroClient{factory: factory}, nil
}

func NewVeleroClientFromRawConfig(rawConfig string) (interfaces.VeleroService, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(rawConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to generate rest.Config from kubeconfig: %w", err)
	}
	return NewVeleroClientFromRestConfig(config)
}

func (v *veleroClient) newKBClient() (kbclient.Client, error) {
	return v.factory.KubebuilderClient()
}

func (v *veleroClient) GetBackups(ctx context.Context, namespace string) ([]velerov1.Backup, error) {
	c, err := v.newKBClient()
	if err != nil {
		return nil, err
	}
	list := &velerov1.BackupList{}
	if err := c.List(ctx, list, kbclient.InNamespace(namespace)); err != nil {
		return nil, utils.WrapK8sError(namespace, err, "backups")
	}

	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (v *veleroClient) GetRestores(ctx context.Context, namespace string) ([]velerov1.Restore, error) {
	c, err := v.newKBClient()
	if err != nil {
		return nil, err
	}
	list := &velerov1.RestoreList{}
	if err := c.List(ctx, list, kbclient.InNamespace(namespace)); err != nil {
		return nil, utils.WrapK8sError(namespace, err, "restores")
	}

	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (v *veleroClient) GetBackupRepositories(ctx context.Context, namespace string) ([]velerov1.BackupRepository, error) {
	c, err := v.newKBClient()
	if err != nil {
		return nil, err
	}
	list := &velerov1.BackupRepositoryList{}
	if err := c.List(ctx, list, kbclient.InNamespace(namespace)); err != nil {
		return nil, utils.WrapK8sError(namespace, err, "backupRepositories")
	}

	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (v *veleroClient) GetBackupStorageLocations(ctx context.Context, namespace string) ([]velerov1.BackupStorageLocation, error) {
	c, err := v.newKBClient()
	if err != nil {
		return nil, err
	}
	list := &velerov1.BackupStorageLocationList{}
	if err := c.List(ctx, list, kbclient.InNamespace(namespace)); err != nil {
		return nil, utils.WrapK8sError(namespace, err, "backupStorageLocations")
	}

	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (v *veleroClient) GetVolumeSnapshotLocations(ctx context.Context, namespace string) ([]velerov1.VolumeSnapshotLocation, error) {
	c, err := v.newKBClient()
	if err != nil {
		return nil, err
	}
	list := &velerov1.VolumeSnapshotLocationList{}
	if err := c.List(ctx, list, kbclient.InNamespace(namespace)); err != nil {
		return nil, utils.WrapK8sError(namespace, err, "volumeSnapshotLocations")
	}

	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (v *veleroClient) GetPodVolumeRestores(ctx context.Context, namespace string) ([]velerov1.PodVolumeRestore, error) {
	c, err := v.newKBClient()
	if err != nil {
		return nil, err
	}
	list := &velerov1.PodVolumeRestoreList{}
	if err := c.List(ctx, list, kbclient.InNamespace(namespace)); err != nil {
		return nil, utils.WrapK8sError(namespace, err, "podVolumeRestores")
	}

	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (v *veleroClient) GetDownloadRequests(ctx context.Context, namespace string) ([]velerov1.DownloadRequest, error) {
	c, err := v.newKBClient()
	if err != nil {
		return nil, err
	}
	list := &velerov1.DownloadRequestList{}
	if err := c.List(ctx, list, kbclient.InNamespace(namespace)); err != nil {
		return nil, utils.WrapK8sError(namespace, err, "downloadRequests")
	}

	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (v *veleroClient) GetDataUploads(ctx context.Context, namespace string) ([]velerov2.DataUpload, error) {
	c, err := v.newKBClient()
	if err != nil {
		return nil, err
	}
	list := &velerov2.DataUploadList{}
	if err := c.List(ctx, list, kbclient.InNamespace(namespace)); err != nil {
		return nil, utils.WrapK8sError(namespace, err, "dataUploads")
	}

	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (v *veleroClient) GetDataDownloads(ctx context.Context, namespace string) ([]velerov2.DataDownload, error) {
	c, err := v.newKBClient()
	if err != nil {
		return nil, err
	}
	list := &velerov2.DataDownloadList{}
	if err := c.List(ctx, list, kbclient.InNamespace(namespace)); err != nil {
		return nil, utils.WrapK8sError(namespace, err, "dataDownloads")
	}

	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (v *veleroClient) GetDeleteBackupRequests(ctx context.Context, namespace string) ([]velerov1.DeleteBackupRequest, error) {
	c, err := v.newKBClient()
	if err != nil {
		return nil, err
	}
	list := &velerov1.DeleteBackupRequestList{}
	if err := c.List(ctx, list, kbclient.InNamespace(namespace)); err != nil {
		return nil, utils.WrapK8sError(namespace, err, "deleteBackupRequests")
	}

	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (v *veleroClient) GetServerStatusRequests(ctx context.Context, namespace string) ([]velerov1.ServerStatusRequest, error) {
	c, err := v.newKBClient()
	if err != nil {
		return nil, err
	}
	list := &velerov1.ServerStatusRequestList{}
	if err := c.List(ctx, list, kbclient.InNamespace(namespace)); err != nil {
		return nil, utils.WrapK8sError(namespace, err, "serverStatusRequests")
	}

	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}
