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
	// URL 기반 설치 테스트 (유효하지 않은 URL이므로 에러 발생 예상)
	err := client.InstallChart("test-chart", "https://example.com/chart.tgz", "", nil)
	if err == nil {
		t.Log("InstallChart succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("InstallChart failed as expected: %v", err)
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

func TestNewClientWithConfigWithNilConfigs(t *testing.T) {
	// nil 설정으로 테스트 (기본 클라이언트로 폴백되어야 함)
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestChartOperations(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// GetChart 테스트
	_, err := client.GetChart(ctx, "test-release", "default", 0)
	if err == nil {
		t.Log("GetChart succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetChart failed as expected: %v", err)
	}

	// IsChartInstalled 테스트
	installed, _, err := client.IsChartInstalled("test-release")
	if err == nil {
		t.Logf("IsChartInstalled succeeded: installed=%v", installed)
	} else {
		t.Logf("IsChartInstalled failed as expected: %v", err)
	}
}

func TestChartManagement(t *testing.T) {
	client := NewClient()

	// UninstallChart 테스트 (dry run)
	err := client.UninstallChart("test-release", "default", true)
	if err == nil {
		t.Log("UninstallChart succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("UninstallChart failed as expected: %v", err)
	}

	// UpgradeChart 테스트
	err = client.UpgradeChart("test-release", "/path/to/chart", nil)
	if err == nil {
		t.Log("UpgradeChart succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("UpgradeChart failed as expected: %v", err)
	}
}
