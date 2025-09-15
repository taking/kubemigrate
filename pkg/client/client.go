package client

import (
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
func NewClient() Client {
	return &client{
		kubernetes: kubernetes.NewClient(),
		helm:       helm.NewClient(),
		velero:     velero.NewClient(),
		minio:      minio.NewClient(),
	}
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

// createClientWithFallback 설정에 따라 클라이언트를 생성하고 실패 시 fallback 사용
func createClientWithFallback[T any, R any](
	config interface{},
	creator func(T) (R, error),
	fallback R,
) R {
	if config == nil {
		return fallback
	}

	if typedConfig, ok := config.(T); ok {
		if client, err := creator(typedConfig); err == nil {
			return client
		}
	}

	return fallback
}

// NewClientWithConfig : 설정을 사용하여 새로운 통합 클라이언트를 생성합니다
func NewClientWithConfig(kubeConfig, helmConfig, veleroConfig, minioConfig interface{}) Client {
	return &client{
		kubernetes: createClientWithFallback[config.KubeConfig, kubernetes.Client]( //nolint:typecheck
			kubeConfig,
			kubernetes.NewClientWithConfig,
			kubernetes.NewClient(),
		),

		helm: createClientWithFallback[config.KubeConfig, helm.Client]( //nolint:typecheck
			helmConfig,
			helm.NewClientWithConfig,
			helm.NewClient(),
		),

		velero: createClientWithFallback[config.VeleroConfig, velero.Client]( //nolint:typecheck
			veleroConfig,
			velero.NewClientWithConfig,
			velero.NewClient(),
		),

		minio: createClientWithFallback[config.MinioConfig, minio.Client]( //nolint:typecheck
			minioConfig,
			minio.NewClientWithConfig,
			minio.NewClient(),
		),
	}
}
