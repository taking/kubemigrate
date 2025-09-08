// Package kubernetes provides a unified client for Kubernetes resource operations.
//
// This package offers a simplified interface for common Kubernetes operations,
// supporting both list and single resource retrieval through unified methods.
//
// Example usage:
//
//	// Create a client
//	client := kubernetes.NewClient()
//
//	// List all pods in default namespace
//	response, err := client.GetPods(ctx, "default", "")
//	if err != nil {
//		return err
//	}
//	podList, ok := response.(*v1.PodList)
//	if !ok {
//		return fmt.Errorf("unexpected response type")
//	}
//
//	// Get a specific pod
//	response, err = client.GetPods(ctx, "default", "my-pod")
//	if err != nil {
//		return err
//	}
//	pod, ok := response.(*v1.Pod)
//	if !ok {
//		return fmt.Errorf("unexpected response type")
//	}
//
// Type Assertion Guide:
// - When name is empty: expect *v1.PodList, *v1.ConfigMapList, *v1.SecretList, *storagev1.StorageClassList
// - When name is provided: expect *v1.Pod, *v1.ConfigMap, *v1.Secret, *storagev1.StorageClass
package kubernetes

import (
	"context"
	"fmt"
	"sort"

	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/validator"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client Kubernetes 클라이언트 인터페이스
type Client interface {
	// 통합 조회 메서드들
	// GetPods returns:
	// - (*v1.PodList, error) when name is empty (list all pods)
	// - (*v1.Pod, error) when name is provided (single pod)
	GetPods(ctx context.Context, namespace, name string) (interface{}, error)

	// GetConfigMaps returns:
	// - (*v1.ConfigMapList, error) when name is empty (list all configmaps)
	// - (*v1.ConfigMap, error) when name is provided (single configmap)
	GetConfigMaps(ctx context.Context, namespace, name string) (interface{}, error)

	// GetSecrets returns:
	// - (*v1.SecretList, error) when name is empty (list all secrets)
	// - (*v1.Secret, error) when name is provided (single secret)
	GetSecrets(ctx context.Context, namespace, name string) (interface{}, error)

	// GetStorageClasses returns:
	// - (*storagev1.StorageClassList, error) when name is empty (list all storage classes)
	// - (*storagev1.StorageClass, error) when name is provided (single storage class)
	GetStorageClasses(ctx context.Context, name string) (interface{}, error)

	// Namespace 관련 (기존 유지)
	GetNamespaces(ctx context.Context) (*v1.NamespaceList, error)
	GetNamespace(ctx context.Context, name string) (*v1.Namespace, error)
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

// GetPods Pod를 조회합니다
// name이 빈 문자열("")이면 목록을 조회하고, 있으면 단일 Pod를 조회합니다
// namespace가 빈 문자열("")이면 모든 네임스페이스의 Pod를 조회합니다
// Returns:
// - (*v1.PodList, error) when name is empty (list all pods)
// - (*v1.Pod, error) when name is provided (single pod)
func (c *client) GetPods(ctx context.Context, namespace, name string) (interface{}, error) {
	if name == "" {
		// 목록 조회
		pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		sort.Slice(pods.Items, func(i, j int) bool {
			return pods.Items[j].CreationTimestamp.Before(&pods.Items[i].CreationTimestamp)
		})

		return pods, nil
	} else {
		// 단일 조회
		return c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	}
}

// GetStorageClasses StorageClass를 조회합니다
// name이 빈 문자열("")이면 목록을 조회하고, 있으면 단일 StorageClass를 조회합니다
// Returns:
// - (*storagev1.StorageClassList, error) when name is empty (list all storage classes)
// - (*storagev1.StorageClass, error) when name is provided (single storage class)
func (c *client) GetStorageClasses(ctx context.Context, name string) (interface{}, error) {
	if name == "" {
		// 목록 조회
		storageClasses, err := c.clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		sort.Slice(storageClasses.Items, func(i, j int) bool {
			return storageClasses.Items[j].CreationTimestamp.Before(&storageClasses.Items[i].CreationTimestamp)
		})
		return storageClasses, nil
	} else {
		// 단일 조회
		return c.clientset.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	}
}

// GetNamespaces Namespace 목록을 조회합니다
func (c *client) GetNamespaces(ctx context.Context) (*v1.NamespaceList, error) {
	return c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
}

// GetNamespace 특정 Namespace를 조회합니다
func (c *client) GetNamespace(ctx context.Context, name string) (*v1.Namespace, error) {
	return c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
}

// GetConfigMaps ConfigMap을 조회합니다
// name이 빈 문자열("")이면 목록을 조회하고, 있으면 단일 ConfigMap을 조회합니다
// namespace가 빈 문자열("")이면 모든 네임스페이스의 ConfigMap을 조회합니다
// Returns:
// - (*v1.ConfigMapList, error) when name is empty (list all configmaps)
// - (*v1.ConfigMap, error) when name is provided (single configmap)
func (c *client) GetConfigMaps(ctx context.Context, namespace, name string) (interface{}, error) {
	if name == "" {
		// 목록 조회
		configMaps, err := c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		sort.Slice(configMaps.Items, func(i, j int) bool {
			return configMaps.Items[j].CreationTimestamp.Before(&configMaps.Items[i].CreationTimestamp)
		})
		return configMaps, nil
	} else {
		// 단일 조회
		return c.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	}
}

// GetSecrets Secret을 조회합니다
// name이 빈 문자열("")이면 목록을 조회하고, 있으면 단일 Secret을 조회합니다
// namespace가 빈 문자열("")이면 모든 네임스페이스의 Secret을 조회합니다
// Returns:
// - (*v1.SecretList, error) when name is empty (list all secrets)
// - (*v1.Secret, error) when name is provided (single secret)
func (c *client) GetSecrets(ctx context.Context, namespace, name string) (interface{}, error) {
	if name == "" {
		// 목록 조회
		secrets, err := c.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		sort.Slice(secrets.Items, func(i, j int) bool {
			return secrets.Items[j].CreationTimestamp.Before(&secrets.Items[i].CreationTimestamp)
		})
		return secrets, nil
	} else {
		// 단일 조회
		return c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	}
}
