package kubernetes

import (
	"context"
	"fmt"

	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/validator"
	"k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client Kubernetes 클라이언트 인터페이스
type Client interface {
	// Pods 관련
	GetPods(ctx context.Context, namespace string) (*v1.PodList, error)
	GetPod(ctx context.Context, namespace, name string) (*v1.Pod, error)

	// StorageClass 관련
	GetStorageClasses(ctx context.Context) (*storagev1.StorageClassList, error)
	GetStorageClass(ctx context.Context, name string) (*storagev1.StorageClass, error)

	// Namespace 관련
	GetNamespaces(ctx context.Context) (*v1.NamespaceList, error)
	GetNamespace(ctx context.Context, name string) (*v1.Namespace, error)

	// ConfigMap 관련
	GetConfigMaps(ctx context.Context, namespace string) (*v1.ConfigMapList, error)
	GetConfigMap(ctx context.Context, namespace, name string) (*v1.ConfigMap, error)

	// Secret 관련
	GetSecrets(ctx context.Context, namespace string) (*v1.SecretList, error)
	GetSecret(ctx context.Context, namespace, name string) (*v1.Secret, error)
}

// client Kubernetes 클라이언트 구현체
type client struct {
	clientset *kubernetes.Clientset
}

// NewClient 새로운 Kubernetes 클라이언트를 생성합니다 (기본 설정)
func NewClient() Client {
	// 기본적으로 in-cluster config를 사용
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		// in-cluster config가 없으면 kubeconfig 파일을 사용
		restConfig, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			panic(err)
		}
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}

	return &client{clientset: clientset}
}

// NewClientWithConfig 설정을 받아서 Kubernetes 클라이언트를 생성합니다
func NewClientWithConfig(cfg config.KubeConfig) (Client, error) {
	var restConfig *rest.Config
	var err error

	if cfg.KubeConfig != "" {
		// 외부 kubeconfig 사용 (base64 디코딩)
		decodedKubeConfig, err := validator.DecodeIfBase64(cfg.KubeConfig)
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

	return &client{clientset: clientset}, nil
}

// GetPods 네임스페이스의 Pod 목록을 조회합니다
func (c *client) GetPods(ctx context.Context, namespace string) (*v1.PodList, error) {
	return c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
}

// GetPod 특정 Pod를 조회합니다
func (c *client) GetPod(ctx context.Context, namespace, name string) (*v1.Pod, error) {
	return c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetStorageClasses StorageClass 목록을 조회합니다
func (c *client) GetStorageClasses(ctx context.Context) (*storagev1.StorageClassList, error) {
	return c.clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
}

// GetStorageClass 특정 StorageClass를 조회합니다
func (c *client) GetStorageClass(ctx context.Context, name string) (*storagev1.StorageClass, error) {
	return c.clientset.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
}

// GetNamespaces Namespace 목록을 조회합니다
func (c *client) GetNamespaces(ctx context.Context) (*v1.NamespaceList, error) {
	return c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
}

// GetNamespace 특정 Namespace를 조회합니다
func (c *client) GetNamespace(ctx context.Context, name string) (*v1.Namespace, error) {
	return c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
}

// GetConfigMaps 네임스페이스의 ConfigMap 목록을 조회합니다
func (c *client) GetConfigMaps(ctx context.Context, namespace string) (*v1.ConfigMapList, error) {
	return c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
}

// GetConfigMap 특정 ConfigMap을 조회합니다
func (c *client) GetConfigMap(ctx context.Context, namespace, name string) (*v1.ConfigMap, error) {
	return c.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetSecrets 네임스페이스의 Secret 목록을 조회합니다
func (c *client) GetSecrets(ctx context.Context, namespace string) (*v1.SecretList, error) {
	return c.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
}

// GetSecret 특정 Secret을 조회합니다
func (c *client) GetSecret(ctx context.Context, namespace, name string) (*v1.Secret, error) {
	return c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
}
