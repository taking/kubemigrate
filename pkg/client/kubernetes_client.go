package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/taking/kubemigrate/pkg/interfaces"
	"github.com/taking/kubemigrate/pkg/models"
	"github.com/taking/kubemigrate/pkg/utils"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"

	kbclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/taking/kubemigrate/pkg/errors"
)

// kubeClient : Kubernetes 클라이언트
type kubeClient struct {
	client  kbclient.Client
	ns      string
	factory *ClientFactory
}

// NewKubernetesClient : Kubernetes 클라이언트 초기화
func NewKubernetesClient(cfg models.KubeConfig) (interfaces.KubernetesClient, error) {
	factory := NewClientFactory()

	restCfg, err := factory.CreateRESTConfig(cfg)
	if err != nil {
		return nil, err
	}

	client, err := kbclient.New(restCfg, kbclient.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ns := getNamespaceOrDefault(cfg.Namespace, "default")

	return &kubeClient{
		client:  client,
		ns:      ns,
		factory: factory,
	}, nil
}

// namespace 기반 ListOption 반환
func (k *kubeClient) listOptions() []kbclient.ListOption {
	if strings.EqualFold(k.ns, "all") { // 대소문자 상관없이 일치할 시
		return nil // 전체 조회
	}
	return []kbclient.ListOption{kbclient.InNamespace(k.ns)}
}

// NewKubeClient : Kubernetes 클라이언트 초기화
func NewKubeClient(cfg models.KubeConfig) (interfaces.KubernetesClient, error) {
	factory := NewClientFactory()

	restCfg, err := factory.CreateRESTConfig(cfg)
	if err != nil {
		return nil, err
	}

	k8sClient, err := kbclient.New(restCfg, kbclient.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ns := getNamespaceOrDefault(cfg.Namespace, "default")

	return &kubeClient{
		client:  k8sClient,
		ns:      ns,
		factory: factory,
	}, nil
}

// HealthCheck : Kubernetes 연결 확인
func (k *kubeClient) HealthCheck(ctx context.Context) error {
	// 5초 제한 타임아웃
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 서버 연결 확인 (Node 목록 조회 시도)
	var nodes v1.NodeList
	return utils.RunWithTimeout(ctx, func() error {
		return k.client.List(ctx, &nodes)
	})
}

func (k *kubeClient) GetPods(ctx context.Context) ([]v1.Pod, error) {
	list := &v1.PodList{}

	// 목록 조회
	if err := k.client.List(ctx, list, k.listOptions()...); err != nil {
		return nil, errors.WrapK8sError(k.ns, err, "pods")
	}

	// 필요 시 ManagedFields 제거
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (k *kubeClient) GetNodes(ctx context.Context) ([]v1.Node, error) {
	list := &v1.NodeList{}

	// 목록 조회
	if err := k.client.List(ctx, list, k.listOptions()...); err != nil {
		return nil, errors.WrapK8sError(k.ns, err, "nodes")
	}

	// 필요 시 ManagedFields 제거
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (k *kubeClient) GetStorageClasses(ctx context.Context) ([]storagev1.StorageClass, error) {
	list := &storagev1.StorageClassList{}

	// 목록 조회
	if err := k.client.List(ctx, list, k.listOptions()...); err != nil {
		return nil, errors.WrapK8sError(k.ns, err, "storageClasses")
	}

	// 필요 시 ManagedFields 제거
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}
