package client

import (
	"fmt"

	"github.com/taking/kubemigrate/pkg/client/helm"
	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/client/velero"
	"github.com/taking/kubemigrate/pkg/config"
)

// Client : 통합 클라이언트 인터페이스
type Client interface {
	Kubernetes() kubernetes.Client
	Helm() helm.Client
	Velero() velero.Client
	Minio() minio.Client
}

// client 통합 클라이언트 구현체
type client struct {
	kubernetes kubernetes.Client
	helm       helm.Client
	velero     velero.Client
	minio      minio.Client
}

// NewClient : 새로운 통합 클라이언트를 생성합니다
func NewClient() (Client, error) {
	kubeClient, err := kubernetes.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	helmClient, err := helm.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create helm client: %w", err)
	}

	minioClient, err := minio.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	veleroClient, err := velero.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create velero client: %w", err)
	}

	return &client{
		kubernetes: kubeClient,
		helm:       helmClient,
		minio:      minioClient,
		velero:     veleroClient,
	}, nil
}

// Kubernetes : Kubernetes 클라이언트를 반환합니다
func (c *client) Kubernetes() kubernetes.Client {
	return c.kubernetes
}

// Helm : Helm 클라이언트를 반환합니다
func (c *client) Helm() helm.Client {
	return c.helm
}

// Velero : Velero 클라이언트를 반환합니다
func (c *client) Velero() velero.Client {
	return c.velero
}

// Minio : MinIO 클라이언트를 반환합니다
func (c *client) Minio() minio.Client {
	return c.minio
}

// createClientWithRetry 설정에 따라 클라이언트를 생성하고 실패 시 fallback 사용
func createClientWithRetry[T any, R any](
	config interface{},
	creator func(T) (R, error),
	fallbackCreator func() (R, error),
) R {
	if config == nil {
		// config가 nil이면 fallback 클라이언트 생성 시도
		if fallback, err := fallbackCreator(); err == nil {
			return fallback
		}
		// fallback도 실패하면 zero value 반환
		var zero R
		return zero
	}

	// 포인터 타입 처리
	if ptr, ok := config.(*T); ok {
		if client, err := creator(*ptr); err == nil {
			return client
		}
	} else if typedConfig, ok := config.(T); ok {
		if client, err := creator(typedConfig); err == nil {
			return client
		}
	}

	// 설정된 클라이언트 생성 실패 시 fallback 클라이언트 생성 시도
	if fallback, err := fallbackCreator(); err == nil {
		return fallback
	}

	// fallback도 실패하면 zero value 반환
	var zero R
	return zero
}

// NewClientWithConfig : 설정을 사용하여 새로운 통합 클라이언트를 생성합니다
func NewClientWithConfig(kubeConfig, helmConfig, veleroConfig, minioConfig interface{}) Client {
	return &client{
		kubernetes: createClientWithRetry[config.KubeConfig, kubernetes.Client]( //nolint:typecheck
			kubeConfig,
			kubernetes.NewClientWithConfig,
			kubernetes.NewClient,
		),

		helm: createClientWithRetry[config.KubeConfig, helm.Client]( //nolint:typecheck
			helmConfig,
			helm.NewClientWithConfig,
			helm.NewClient,
		),

		velero: createClientWithRetry[config.VeleroConfig, velero.Client]( //nolint:typecheck
			veleroConfig,
			velero.NewClientWithConfig,
			velero.NewClient,
		),

		minio: createClientWithRetry[config.MinioConfig, minio.Client]( //nolint:typecheck
			minioConfig,
			minio.NewClientWithConfig,
			minio.NewClient,
		),
	}
}
