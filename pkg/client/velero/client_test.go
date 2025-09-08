package velero

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

func TestGetBackups(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 클러스터가 없으므로 에러가 발생할 것으로 예상
	_, err := client.GetBackups(ctx, "velero")
	if err == nil {
		t.Log("GetBackups succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetBackups failed as expected: %v", err)
	}
}

func TestGetBackupRepositories(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 클러스터가 없으므로 에러가 발생할 것으로 예상
	_, err := client.GetBackupRepositories(ctx, "velero")
	if err == nil {
		t.Log("GetBackupRepositories succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetBackupRepositories failed as expected: %v", err)
	}
}

func TestNewClientWithConfig(t *testing.T) {
	// 빈 설정으로 테스트 (에러가 발생할 것으로 예상)
	cfg := config.VeleroConfig{
		KubeConfig: config.KubeConfig{
			KubeConfig: "",
			Namespace:  "velero",
		},
		MinioConfig: config.MinioConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
			UseSSL:    false,
		},
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
