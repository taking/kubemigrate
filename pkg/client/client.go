package client

import (
	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/pkg/client/helm"
	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/client/velero"
)

// Client 통합 클라이언트 인터페이스
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

// NewClient 새로운 통합 클라이언트를 생성합니다
func NewClient() Client {
	return &client{
		kubernetes: kubernetes.NewClient(),
		helm:       helm.NewClient(),
		velero:     velero.NewClient(),
		minio:      minio.NewClient(),
	}
}

// Kubernetes Kubernetes 클라이언트를 반환합니다
func (c *client) Kubernetes() kubernetes.Client {
	return c.kubernetes
}

// Helm Helm 클라이언트를 반환합니다
func (c *client) Helm() helm.Client {
	return c.helm
}

// Velero Velero 클라이언트를 반환합니다
func (c *client) Velero() velero.Client {
	return c.velero
}

// Minio MinIO 클라이언트를 반환합니다
func (c *client) Minio() minio.Client {
	return c.minio
}

// NewClientWithConfig 설정을 사용하여 새로운 통합 클라이언트를 생성합니다
func NewClientWithConfig(kubeConfig, helmConfig, veleroConfig, minioConfig interface{}) Client {
	// 타입 어설션 및 에러 처리
	var k8sClient kubernetes.Client
	var helmClient helm.Client
	var veleroClient velero.Client
	var minioClient minio.Client

	if kubeCfg, ok := kubeConfig.(config.KubeConfig); ok {
		if client, err := kubernetes.NewClientWithConfig(kubeCfg); err == nil {
			k8sClient = client
		} else {
			k8sClient = kubernetes.NewClient() // 기본 클라이언트 사용
		}
	} else {
		k8sClient = kubernetes.NewClient() // 기본 클라이언트 사용
	}

	if helmCfg, ok := helmConfig.(config.HelmConfig); ok {
		if client, err := helm.NewClientWithConfig(helmCfg); err == nil {
			helmClient = client
		} else {
			helmClient = helm.NewClient() // 기본 클라이언트 사용
		}
	} else {
		helmClient = helm.NewClient() // 기본 클라이언트 사용
	}

	if veleroCfg, ok := veleroConfig.(config.VeleroConfig); ok {
		if client, err := velero.NewClientWithConfig(veleroCfg); err == nil {
			veleroClient = client
		} else {
			veleroClient = velero.NewClient() // 기본 클라이언트 사용
		}
	} else {
		veleroClient = velero.NewClient() // 기본 클라이언트 사용
	}

	if minioCfg, ok := minioConfig.(config.MinioConfig); ok {
		if client, err := minio.NewClientWithConfig(minioCfg); err == nil {
			minioClient = client
		} else {
			minioClient = minio.NewClient() // 기본 클라이언트 사용
		}
	} else {
		minioClient = minio.NewClient() // 기본 클라이언트 사용
	}

	return &client{
		kubernetes: k8sClient,
		helm:       helmClient,
		velero:     veleroClient,
		minio:      minioClient,
	}
}
