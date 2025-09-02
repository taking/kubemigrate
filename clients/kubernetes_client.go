package clients

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/client-go/rest"
	"taking.kr/velero/models"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	kbclient "sigs.k8s.io/controller-runtime/pkg/client"

	"taking.kr/velero/utils"
)

type KubeClient struct {
	client kbclient.Client
	ns     string
}

// NewKubeClient : Kubernetes 클라이언트 초기화
func NewKubeClient(cfg models.KubeConfig) (*KubeClient, error) {
	var restCfg *rest.Config
	var err error

	if cfg.KubeConfig != "" {
		restCfg, err = clientcmd.RESTConfigFromKubeConfig([]byte(cfg.KubeConfig))
		if err != nil {
			return nil, fmt.Errorf("❌ failed to parse kubeconfig: %w", err)
		}
	}

	k8sClient, err := kbclient.New(restCfg, kbclient.Options{})
	if err != nil {
		return nil, fmt.Errorf("❌ failed to create kubernetes client: %w", err)
	}

	// 네임스페이스 없을 시, "default"로 설정
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	return &KubeClient{
		client: k8sClient,
		ns:     cfg.Namespace,
	}, nil
}

// HealthCheck : Kubernetes 연결 확인
func (k *KubeClient) HealthCheck(ctx context.Context) error {
	// 5초 제한 타임아웃
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 서버 연결 확인 (백업 목록 조회 시도)
	var pods v1.PodList
	if err := k.client.List(ctx, &pods, kbclient.InNamespace(k.ns)); err != nil {
		return fmt.Errorf("❌ failed to kubernetes health check: %w", err)
	}
	return nil
}

func (k *KubeClient) GetPods(ctx context.Context) ([]v1.Pod, error) {
	list := &v1.PodList{}

	// 목록 조회
	if err := k.client.List(ctx, list); err != nil {
		return nil, utils.WrapK8sError("", err, "pods")
	}

	// 필요 시 ManagedFields 제거
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}

func (k *KubeClient) GetStorageClasses(ctx context.Context) ([]storagev1.StorageClass, error) {
	list := &storagev1.StorageClassList{}

	// 목록 조회
	if err := k.client.List(ctx, list); err != nil {
		return nil, utils.WrapK8sError("", err, "storageClasses")
	}

	// 필요 시 ManagedFields 제거
	for i := range list.Items {
		utils.StripManagedFields(&list.Items[i])
	}

	return list.Items, nil
}
