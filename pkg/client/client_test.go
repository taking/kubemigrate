package client

import (
	"testing"

	"github.com/taking/kubemigrate/internal/config"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	// 각 클라이언트가 nil이 아닌지 확인
	if client.Kubernetes() == nil {
		t.Error("Kubernetes client is nil")
	}
	if client.Helm() == nil {
		t.Error("Helm client is nil")
	}
	if client.Velero() == nil {
		t.Error("Velero client is nil")
	}
	if client.Minio() == nil {
		t.Error("MinIO client is nil")
	}
}

func TestNewClientWithConfig(t *testing.T) {
	// 빈 설정으로 테스트
	kubeConfig := config.KubeConfig{
		KubeConfig: "",
		Namespace:  "default",
	}

	helmConfig := config.HelmConfig{
		KubeConfig: kubeConfig,
	}

	veleroConfig := config.VeleroConfig{
		KubeConfig: kubeConfig,
		MinioConfig: config.MinioConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
			UseSSL:    false,
		},
	}

	minioConfig := config.MinioConfig{
		Endpoint:  "localhost:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		UseSSL:    false,
	}

	client := NewClientWithConfig(kubeConfig, helmConfig, veleroConfig, minioConfig)
	if client == nil {
		t.Fatal("NewClientWithConfig() returned nil")
	}

	// 각 클라이언트가 nil이 아닌지 확인
	if client.Kubernetes() == nil {
		t.Error("Kubernetes client is nil")
	}
	if client.Helm() == nil {
		t.Error("Helm client is nil")
	}
	if client.Velero() == nil {
		t.Error("Velero client is nil")
	}
	if client.Minio() == nil {
		t.Error("MinIO client is nil")
	}
}

func TestNewClientWithConfigWithNilConfigs(t *testing.T) {
	// nil 설정으로 테스트 (기본 클라이언트로 폴백되어야 함)
	client := NewClientWithConfig(nil, nil, nil, nil)
	if client == nil {
		t.Fatal("NewClientWithConfig() returned nil")
	}

	// 각 클라이언트가 nil이 아닌지 확인
	if client.Kubernetes() == nil {
		t.Error("Kubernetes client is nil")
	}
	if client.Helm() == nil {
		t.Error("Helm client is nil")
	}
	if client.Velero() == nil {
		t.Error("Velero client is nil")
	}
	if client.Minio() == nil {
		t.Error("MinIO client is nil")
	}
}
