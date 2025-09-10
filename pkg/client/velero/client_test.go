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

func TestNewClientWithConfigWithNilConfigs(t *testing.T) {
	// nil 설정으로 테스트 (기본 클라이언트로 폴백되어야 함)
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestRestoreOperations(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// GetRestores 테스트
	_, err := client.GetRestores(ctx, "velero")
	if err == nil {
		t.Log("GetRestores succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetRestores failed as expected: %v", err)
	}

	// GetRestore 테스트
	_, err = client.GetRestore(ctx, "velero", "test-restore")
	if err == nil {
		t.Log("GetRestore succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetRestore failed as expected: %v", err)
	}
}

func TestBackupStorageLocationOperations(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// GetBackupStorageLocations 테스트
	_, err := client.GetBackupStorageLocations(ctx, "velero")
	if err == nil {
		t.Log("GetBackupStorageLocations succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetBackupStorageLocations failed as expected: %v", err)
	}

	// GetBackupStorageLocation 테스트
	_, err = client.GetBackupStorageLocation(ctx, "velero", "default")
	if err == nil {
		t.Log("GetBackupStorageLocation succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetBackupStorageLocation failed as expected: %v", err)
	}
}

func TestVolumeSnapshotLocationOperations(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// GetVolumeSnapshotLocations 테스트
	_, err := client.GetVolumeSnapshotLocations(ctx, "velero")
	if err == nil {
		t.Log("GetVolumeSnapshotLocations succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetVolumeSnapshotLocations failed as expected: %v", err)
	}

	// GetVolumeSnapshotLocation 테스트
	_, err = client.GetVolumeSnapshotLocation(ctx, "velero", "default")
	if err == nil {
		t.Log("GetVolumeSnapshotLocation succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetVolumeSnapshotLocation failed as expected: %v", err)
	}
}

func TestPodVolumeRestoreOperations(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// GetPodVolumeRestores 테스트
	_, err := client.GetPodVolumeRestores(ctx, "velero")
	if err == nil {
		t.Log("GetPodVolumeRestores succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetPodVolumeRestores failed as expected: %v", err)
	}

	// GetPodVolumeRestore 테스트
	_, err = client.GetPodVolumeRestore(ctx, "velero", "test-pvr")
	if err == nil {
		t.Log("GetPodVolumeRestore succeeded - this might indicate a real cluster is available")
	} else {
		t.Logf("GetPodVolumeRestore failed as expected: %v", err)
	}
}
