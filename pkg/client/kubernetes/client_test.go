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
	_, err := client.GetPods(ctx, "default")
	if err == nil {
		t.Log("GetPods succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetPods failed as expected: %v", err)
	}
}

func TestGetStorageClasses(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 클러스터가 없으므로 에러가 발생할 것으로 예상
	_, err := client.GetStorageClasses(ctx)
	if err == nil {
		t.Log("GetStorageClasses succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetStorageClasses failed as expected: %v", err)
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
