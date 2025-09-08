package kubernetes

import (
	"context"
	"testing"

	"github.com/taking/kubemigrate/internal/config"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestGetPods(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 클러스터가 없으므로 에러가 발생할 것으로 예상
	// List pods
	_, err := client.GetPods(ctx, "default", "")
	if err == nil {
		t.Log("GetPods (list) succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetPods (list) failed as expected: %v", err)
	}

	// Get single pod
	_, err = client.GetPods(ctx, "default", "test-pod")
	if err == nil {
		t.Log("GetPods (single) succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetPods (single) failed as expected: %v", err)
	}
}

func TestGetStorageClasses(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 클러스터가 없으므로 에러가 발생할 것으로 예상
	// List storage classes
	_, err := client.GetStorageClasses(ctx, "")
	if err == nil {
		t.Log("GetStorageClasses (list) succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetStorageClasses (list) failed as expected: %v", err)
	}

	// Get single storage class
	_, err = client.GetStorageClasses(ctx, "test-storage-class")
	if err == nil {
		t.Log("GetStorageClasses (single) succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetStorageClasses (single) failed as expected: %v", err)
	}
}

func TestNewClientWithConfig(t *testing.T) {
	// 빈 설정으로 테스트 (에러가 발생할 것으로 예상)
	cfg := config.KubeConfig{
		KubeConfig: "",
		Namespace:  "default",
	}

	client, err := NewClientWithConfig(cfg)
	if err == nil {
		t.Log("NewClientWithConfig succeeded - this might indicate a real cluster is available")
		if client == nil {
			t.Fatal("NewClientWithConfig returned nil client")
		}
	} else {
		t.Logf("NewClientWithConfig failed as expected: %v", err)
	}
}

func TestGetConfigMaps(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 클러스터가 없으므로 에러가 발생할 것으로 예상
	// List configmaps
	_, err := client.GetConfigMaps(ctx, "default", "")
	if err == nil {
		t.Log("GetConfigMaps (list) succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetConfigMaps (list) failed as expected: %v", err)
	}

	// Get single configmap
	_, err = client.GetConfigMaps(ctx, "default", "test-configmap")
	if err == nil {
		t.Log("GetConfigMaps (single) succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetConfigMaps (single) failed as expected: %v", err)
	}
}

func TestGetSecrets(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 클러스터가 없으므로 에러가 발생할 것으로 예상
	// List secrets
	_, err := client.GetSecrets(ctx, "default", "")
	if err == nil {
		t.Log("GetSecrets (list) succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetSecrets (list) failed as expected: %v", err)
	}

	// Get single secret
	_, err = client.GetSecrets(ctx, "default", "test-secret")
	if err == nil {
		t.Log("GetSecrets (single) succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetSecrets (single) failed as expected: %v", err)
	}
}
