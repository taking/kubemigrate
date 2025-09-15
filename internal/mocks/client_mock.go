package mocks

import (
	"context"
	"io"
	"time"

	miniosdk "github.com/minio/minio-go/v7"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/client/helm"
	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/client/velero"
	"github.com/taking/kubemigrate/pkg/config"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MockClient : 테스트용 Mock 클라이언트
type MockClient struct{}

// NewMockClient : Mock 클라이언트 생성
func NewMockClient() *MockClient {
	return &MockClient{}
}

// Kubernetes : Mock Kubernetes 클라이언트 반환
func (m *MockClient) Kubernetes() kubernetes.Client {
	return &MockKubernetesClient{}
}

// Helm : Mock Helm 클라이언트 반환
func (m *MockClient) Helm() helm.Client {
	return &MockHelmClient{}
}

// Minio : Mock MinIO 클라이언트 반환
func (m *MockClient) Minio() minio.Client {
	return &MockMinioClient{}
}

// Velero : Mock Velero 클라이언트 반환
func (m *MockClient) Velero() velero.Client {
	return &MockVeleroClient{}
}

// MockKubernetesClient : Mock Kubernetes 클라이언트
type MockKubernetesClient struct{}

func (m *MockKubernetesClient) GetNamespaces(ctx context.Context) (*v1.NamespaceList, error) {
	return &v1.NamespaceList{
		Items: []v1.Namespace{
			{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetNamespace(ctx context.Context, name string) (*v1.Namespace, error) {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}, nil
}

func (m *MockKubernetesClient) GetStorageClasses(ctx context.Context, name string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-storageclass"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetPods(ctx context.Context, namespace, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-pod"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetConfigMaps(ctx context.Context, namespace, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-configmap"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetSecrets(ctx context.Context, namespace, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-secret"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetServices(ctx context.Context, namespace, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-service"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetDeployments(ctx context.Context, namespace, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-deployment"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetStatefulSets(ctx context.Context, namespace, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-statefulset"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetDaemonSets(ctx context.Context, namespace, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-daemonset"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetJobs(ctx context.Context, namespace, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-job"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetCronJobs(ctx context.Context, namespace, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-cronjob"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetIngresses(ctx context.Context, namespace, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-ingress"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetPersistentVolumes(ctx context.Context, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-pv"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetPersistentVolumeClaims(ctx context.Context, namespace, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-pvc"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetNodes(ctx context.Context, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"metadata": map[string]interface{}{"name": "test-node"}},
		},
	}, nil
}

func (m *MockKubernetesClient) GetResource(ctx context.Context, resourceType, namespace, name, labelSelector string) (interface{}, error) {
	return map[string]interface{}{
		"metadata": map[string]interface{}{"name": "test-" + resourceType},
	}, nil
}

// MockHelmClient : Mock Helm 클라이언트
type MockHelmClient struct{}

func (m *MockHelmClient) GetCharts(ctx context.Context, namespace string) ([]*release.Release, error) {
	return []*release.Release{
		{Name: "test-chart", Namespace: namespace, Info: &release.Info{Status: "deployed"}},
	}, nil
}

func (m *MockHelmClient) GetChart(ctx context.Context, releaseName, namespace string, releaseVersion int) (*release.Release, error) {
	return &release.Release{
		Name:      releaseName,
		Namespace: namespace,
		Version:   releaseVersion,
		Info:      &release.Info{Status: "deployed"},
	}, nil
}

func (m *MockHelmClient) IsChartInstalled(releaseName string) (bool, *release.Release, error) {
	return true, &release.Release{Name: releaseName}, nil
}

func (m *MockHelmClient) InstallChart(releaseName, chartURL, version string, values map[string]interface{}) error {
	return nil
}

func (m *MockHelmClient) UninstallChart(releaseName, namespace string, dryRun bool) error {
	return nil
}

func (m *MockHelmClient) UpgradeChart(releaseName, chartPath string, values map[string]interface{}) error {
	return nil
}

// MockMinioClient : Mock MinIO 클라이언트
type MockMinioClient struct{}

func (m *MockMinioClient) ListBuckets(ctx context.Context) (interface{}, error) {
	return []miniosdk.BucketInfo{
		{Name: "test-bucket", CreationDate: time.Now()},
	}, nil
}

func (m *MockMinioClient) CreateBucket(ctx context.Context, bucketName string) error {
	return nil
}

func (m *MockMinioClient) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	return nil
}

func (m *MockMinioClient) CreateBucketIfNotExists(ctx context.Context, bucketName string) error {
	return nil
}

func (m *MockMinioClient) DeleteBucket(ctx context.Context, bucketName string) error {
	return nil
}

func (m *MockMinioClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return true, nil
}

func (m *MockMinioClient) ListObjects(ctx context.Context, bucketName string) (interface{}, error) {
	return []miniosdk.ObjectInfo{
		{Key: "test-object", Size: 1024},
	}, nil
}

func (m *MockMinioClient) GetObject(ctx context.Context, bucketName, objectName string) (interface{}, error) {
	return map[string]interface{}{
		"key":  objectName,
		"size": 1024,
		"data": []byte("test data"),
	}, nil
}

func (m *MockMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64) (interface{}, error) {
	return map[string]interface{}{
		"bucket": bucketName,
		"object": objectName,
		"size":   objectSize,
	}, nil
}

func (m *MockMinioClient) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	return nil
}

func (m *MockMinioClient) StatObject(ctx context.Context, bucketName, objectName string) (interface{}, error) {
	return map[string]interface{}{
		"key":  objectName,
		"size": 1024,
	}, nil
}

func (m *MockMinioClient) CopyObject(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string) (interface{}, error) {
	return map[string]interface{}{
		"srcBucket": srcBucket,
		"srcObject": srcObject,
		"dstBucket": dstBucket,
		"dstObject": dstObject,
	}, nil
}

func (m *MockMinioClient) PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry int) (string, error) {
	return "https://test.example.com/presigned-url", nil
}

func (m *MockMinioClient) PresignedPutObject(ctx context.Context, bucketName, objectName string, expiry int) (string, error) {
	return "https://test.example.com/presigned-put-url", nil
}

// MockVeleroClient : Mock Velero 클라이언트
type MockVeleroClient struct{}

func (m *MockVeleroClient) GetBackups(ctx context.Context, namespace string) ([]velerov1.Backup, error) {
	return []velerov1.Backup{
		{ObjectMeta: metav1.ObjectMeta{Name: "test-backup", Namespace: namespace}},
	}, nil
}

func (m *MockVeleroClient) GetBackup(ctx context.Context, namespace, name string) (*velerov1.Backup, error) {
	return &velerov1.Backup{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}, nil
}

func (m *MockVeleroClient) CreateBackup(ctx context.Context, namespace string, backup *velerov1.Backup) error {
	return nil
}

func (m *MockVeleroClient) DeleteBackup(ctx context.Context, namespace, name string) error {
	return nil
}

func (m *MockVeleroClient) GetRestores(ctx context.Context, namespace string) ([]velerov1.Restore, error) {
	return []velerov1.Restore{
		{ObjectMeta: metav1.ObjectMeta{Name: "test-restore", Namespace: namespace}},
	}, nil
}

func (m *MockVeleroClient) GetRestore(ctx context.Context, namespace, name string) (*velerov1.Restore, error) {
	return &velerov1.Restore{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}, nil
}

func (m *MockVeleroClient) CreateRestore(ctx context.Context, namespace string, restore *velerov1.Restore) error {
	return nil
}

func (m *MockVeleroClient) DeleteRestore(ctx context.Context, namespace, name string) error {
	return nil
}

func (m *MockVeleroClient) GetBackupRepositories(ctx context.Context, namespace string) ([]velerov1.BackupRepository, error) {
	return []velerov1.BackupRepository{
		{ObjectMeta: metav1.ObjectMeta{Name: "test-repo", Namespace: namespace}},
	}, nil
}

func (m *MockVeleroClient) GetBackupRepository(ctx context.Context, namespace, name string) (*velerov1.BackupRepository, error) {
	return &velerov1.BackupRepository{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}, nil
}

func (m *MockVeleroClient) GetBackupStorageLocations(ctx context.Context, namespace string) ([]velerov1.BackupStorageLocation, error) {
	return []velerov1.BackupStorageLocation{
		{ObjectMeta: metav1.ObjectMeta{Name: "test-bsl", Namespace: namespace}},
	}, nil
}

func (m *MockVeleroClient) GetBackupStorageLocation(ctx context.Context, namespace, name string) (*velerov1.BackupStorageLocation, error) {
	return &velerov1.BackupStorageLocation{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}, nil
}

func (m *MockVeleroClient) GetVolumeSnapshotLocations(ctx context.Context, namespace string) ([]velerov1.VolumeSnapshotLocation, error) {
	return []velerov1.VolumeSnapshotLocation{
		{ObjectMeta: metav1.ObjectMeta{Name: "test-vsl", Namespace: namespace}},
	}, nil
}

func (m *MockVeleroClient) GetVolumeSnapshotLocation(ctx context.Context, namespace, name string) (*velerov1.VolumeSnapshotLocation, error) {
	return &velerov1.VolumeSnapshotLocation{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}, nil
}

func (m *MockVeleroClient) GetPodVolumeRestores(ctx context.Context, namespace string) ([]velerov1.PodVolumeRestore, error) {
	return []velerov1.PodVolumeRestore{
		{ObjectMeta: metav1.ObjectMeta{Name: "test-pvr", Namespace: namespace}},
	}, nil
}

func (m *MockVeleroClient) GetPodVolumeRestore(ctx context.Context, namespace, name string) (*velerov1.PodVolumeRestore, error) {
	return &velerov1.PodVolumeRestore{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}, nil
}

// MockClientFactory : Mock 클라이언트 팩토리
type MockClientFactory struct{}

// NewMockClientFactory : Mock 클라이언트 팩토리 생성
func NewMockClientFactory() *MockClientFactory {
	return &MockClientFactory{}
}

// CreateClient : Mock 클라이언트 생성
func (f *MockClientFactory) CreateClient(kubeConfig *config.KubeConfig, helmConfig *config.KubeConfig, veleroConfig *config.VeleroConfig, minioConfig *config.MinioConfig) client.Client {
	return NewMockClient()
}
