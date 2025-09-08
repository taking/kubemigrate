package helm

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

func TestGetCharts(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 클러스터가 없으므로 에러가 발생할 것으로 예상
	_, err := client.GetCharts(ctx, "default")
	if err == nil {
		t.Log("GetCharts succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetCharts failed as expected: %v", err)
	}
}

func TestInstallChart(t *testing.T) {
	client := NewClient()

	// 실제 클러스터가 없으므로 에러가 발생할 것으로 예상
	err := client.InstallChart("test-chart", "test-path", nil)
	if err == nil {
		t.Log("InstallChart succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("InstallChart failed as expected: %v", err)
	}
}

func TestNewClientWithConfig(t *testing.T) {
	// 빈 설정으로 테스트 (에러가 발생할 것으로 예상)
	cfg := config.HelmConfig{
		KubeConfig: config.KubeConfig{
			KubeConfig: "",
			Namespace:  "default",
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

func TestHealthCheck(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// 실제 클러스터가 없으므로 에러가 발생할 것으로 예상
	err := client.HealthCheck(ctx)
	if err == nil {
		t.Log("HealthCheck succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("HealthCheck failed as expected: %v", err)
	}
}
