package clients

import (
	"context"
	"fmt"
	"taking.kr/velero/interfaces"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type kubeClient struct {
	client dynamic.Interface
}

func NewKubeClientFromRestConfig(config *rest.Config) (interfaces.KubeService, error) {
	cl, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}
	return &kubeClient{client: cl}, nil
}

func NewKubeClientFromRawConfig(rawConfig string) (interfaces.KubeService, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(rawConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create rest.Config from raw kubeconfig: %w", err)
	}
	return NewKubeClientFromRestConfig(config)
}

func (k *kubeClient) GetResources(ctx context.Context, resource schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	obj, err := k.client.Resource(resource).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}
	return obj, nil
}
