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

	"github.com/taking/kubemigrate/internal/validator"
	"github.com/taking/kubemigrate/pkg/config"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client : Kubernetes 클라이언트 인터페이스
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

	// CreateNamespace creates a new namespace
	CreateNamespace(ctx context.Context, namespace *v1.Namespace) (interface{}, error)

	// DeleteSecret deletes a secret
	DeleteSecret(ctx context.Context, namespace, name string) error

	// DeleteCRD deletes a custom resource definition
	DeleteCRD(ctx context.Context, name string) error

	// DeleteNamespace deletes a namespace
	DeleteNamespace(ctx context.Context, name string) error

	// GetNamespaces returns:
	// - (*v1.NamespaceList, error) when name is empty (list all namespaces)
	// - (*v1.Namespace, error) when name is provided (single namespace)
	GetNamespaces(ctx context.Context, name string) (interface{}, error)

	// Secret 생성
	CreateSecret(ctx context.Context, namespace, name string, data map[string]string) (*v1.Secret, error)
}

// client Kubernetes 클라이언트 구현체
type client struct {
	clientset    *kubernetes.Clientset
	crdClientset *apiextensionsclientset.Clientset
}

// NewClient : 새로운 Kubernetes 클라이언트를 생성합니다 (기본 설정)
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

	// Rate limiter 설정 (QPS: 50, Burst: 100)
	restConfig.QPS = 50
	restConfig.Burst = 100

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	crdClientset, err := apiextensionsclientset.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create crd client: %w", err)
	}

	return &client{
		clientset:    clientset,
		crdClientset: crdClientset,
	}, nil
}

// NewClientWithConfig : 설정을 받아서 Kubernetes 클라이언트를 생성합니다
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

	// Rate limiter 설정 (QPS: 50, Burst: 100)
	restConfig.QPS = 50
	restConfig.Burst = 100

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	crdClientset, err := apiextensionsclientset.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create crd client: %w", err)
	}

	return &client{
		clientset:    clientset,
		crdClientset: crdClientset,
	}, nil
}

// stripManagedFieldsFromList : 리소스 목록에서 managedFields 제거
func stripManagedFieldsFromList(items interface{}) {
	switch list := items.(type) {
	case *v1.PodList:
		for i := range list.Items {
			list.Items[i].SetManagedFields(nil)
		}
	case *v1.ConfigMapList:
		for i := range list.Items {
			list.Items[i].SetManagedFields(nil)
		}
	case *v1.SecretList:
		for i := range list.Items {
			list.Items[i].SetManagedFields(nil)
		}
	case *v1.NamespaceList:
		for i := range list.Items {
			list.Items[i].SetManagedFields(nil)
		}
	case *storagev1.StorageClassList:
		for i := range list.Items {
			list.Items[i].SetManagedFields(nil)
		}
	}
}

// stripManagedFieldsFromSingle : 단일 리소스에서 managedFields 제거
func stripManagedFieldsFromSingle(obj interface{}) {
	switch resource := obj.(type) {
	case *v1.Pod:
		resource.SetManagedFields(nil)
	case *v1.ConfigMap:
		resource.SetManagedFields(nil)
	case *v1.Secret:
		resource.SetManagedFields(nil)
	case *v1.Namespace:
		resource.SetManagedFields(nil)
	case *storagev1.StorageClass:
		resource.SetManagedFields(nil)
	}
}

// GetPods : Pod를 조회합니다
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

		// managedFields 제거
		stripManagedFieldsFromList(pods)

		return pods, nil
	} else {
		// 단일 조회
		pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// managedFields 제거
		stripManagedFieldsFromSingle(pod)

		return pod, nil
	}
}

// GetStorageClasses : StorageClass를 조회합니다
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

		// managedFields 제거
		stripManagedFieldsFromList(storageClasses)

		return storageClasses, nil
	} else {
		// 단일 조회
		storageClass, err := c.clientset.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// managedFields 제거
		stripManagedFieldsFromSingle(storageClass)

		return storageClass, nil
	}
}

// GetNamespaces : Namespace를 조회합니다
// name이 빈 문자열("")이면 목록을 조회하고, 있으면 단일 Namespace를 조회합니다
func (c *client) GetNamespaces(ctx context.Context, name string) (interface{}, error) {
	if name == "" {
		// 목록 조회
		namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		sort.Slice(namespaces.Items, func(i, j int) bool {
			return namespaces.Items[j].CreationTimestamp.Before(&namespaces.Items[i].CreationTimestamp)
		})

		// managedFields 제거
		stripManagedFieldsFromList(namespaces)

		return namespaces, nil
	} else {
		// 단일 조회
		namespace, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// managedFields 제거
		stripManagedFieldsFromSingle(namespace)

		return namespace, nil
	}
}

// GetConfigMaps : ConfigMap을 조회합니다
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

		// managedFields 제거
		stripManagedFieldsFromList(configMaps)

		return configMaps, nil
	} else {
		// 단일 조회
		configMap, err := c.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// managedFields 제거
		stripManagedFieldsFromSingle(configMap)

		return configMap, nil
	}
}

// GetSecrets : Secret을 조회합니다
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

		// managedFields 제거
		stripManagedFieldsFromList(secrets)

		return secrets, nil
	} else {
		// 단일 조회
		secret, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		// managedFields 제거
		stripManagedFieldsFromSingle(secret)

		return secret, nil
	}
}

// CreateSecret : Secret 생성
func (c *client) CreateSecret(ctx context.Context, namespace, name string, data map[string]string) (*v1.Secret, error) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: v1.SecretTypeOpaque,
		Data: make(map[string][]byte),
	}

	// string 데이터를 []byte로 변환
	for key, value := range data {
		secret.Data[key] = []byte(value)
	}

	return c.clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
}

// CreateNamespace : Namespace 생성
func (c *client) CreateNamespace(ctx context.Context, namespace *v1.Namespace) (interface{}, error) {
	return c.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
}

// DeleteCRD : Custom Resource Definition 삭제
func (c *client) DeleteCRD(ctx context.Context, name string) error {
	return c.crdClientset.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, name, metav1.DeleteOptions{})
}

// DeleteNamespace : Namespace 삭제
func (c *client) DeleteNamespace(ctx context.Context, name string) error {
	return c.clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}

// DeleteSecret : Secret 삭제
func (c *client) DeleteSecret(ctx context.Context, namespace, name string) error {
	return c.clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
