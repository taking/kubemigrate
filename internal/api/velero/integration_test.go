package velero

import (
	"context"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/mocks"
	"github.com/taking/kubemigrate/pkg/config"
	"github.com/taking/kubemigrate/pkg/utils"
)

// TestVeleroInstallationIntegration : Velero 설치 통합 테스트
func TestVeleroInstallationIntegration(t *testing.T) {
	// Mock 클라이언트 설정
	mockClient := mocks.NewMockClient()

	// BaseHandler 생성 (Mock 사용)
	baseHandler := handler.NewBaseHandlerWithMock(nil)
	baseHandler.UseMockClient(true)

	// Velero Handler 생성
	veleroHandler := NewHandler(baseHandler)

	// 테스트 설정
	veleroConfig := config.VeleroConfig{
		KubeConfig: config.KubeConfig{
			KubeConfig: "apiVersion: v1\nkind: Config",
			Namespace:  "velero",
		},
		MinioConfig: config.MinioConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
			UseSSL:    false,
		},
	}

	// Echo 컨텍스트 생성
	e := echo.New()
	req := e.NewRequest(nil, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := e.NewResponse(nil, nil)
	c := e.NewContext(req, rec)

	// JSON 바인딩을 위한 요청 본문 설정
	c.SetRequest(req)

	// 테스트 실행
	err := veleroHandler.InstallVeleroWithMinIO(c)

	// 결과 검증
	assert.NoError(t, err)
}

// TestVeleroHealthCheckIntegration : Velero HealthCheck 통합 테스트
func TestVeleroHealthCheckIntegration(t *testing.T) {
	// Mock 클라이언트 설정
	mockClient := mocks.NewMockClient()

	// BaseHandler 생성 (Mock 사용)
	baseHandler := handler.NewBaseHandlerWithMock(nil)
	baseHandler.UseMockClient(true)

	// Velero Handler 생성
	veleroHandler := NewHandler(baseHandler)

	// Echo 컨텍스트 생성
	e := echo.New()
	req := e.NewRequest(nil, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := e.NewResponse(nil, nil)
	c := e.NewContext(req, rec)

	// JSON 바인딩을 위한 요청 본문 설정
	c.SetRequest(req)

	// 테스트 실행
	err := veleroHandler.HealthCheck(c)

	// 결과 검증
	assert.NoError(t, err)
}

// TestVeleroBackupIntegration : Velero Backup 통합 테스트
func TestVeleroBackupIntegration(t *testing.T) {
	// Mock 클라이언트 설정
	mockClient := mocks.NewMockClient()

	// BaseHandler 생성 (Mock 사용)
	baseHandler := handler.NewBaseHandlerWithMock(nil)
	baseHandler.UseMockClient(true)

	// Velero Handler 생성
	veleroHandler := NewHandler(baseHandler)

	// Echo 컨텍스트 생성
	e := echo.New()
	req := e.NewRequest(nil, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := e.NewResponse(nil, nil)
	c := e.NewContext(req, rec)

	// JSON 바인딩을 위한 요청 본문 설정
	c.SetRequest(req)

	// 테스트 실행
	err := veleroHandler.GetBackups(c)

	// 결과 검증
	assert.NoError(t, err)
}

// TestVeleroRestoreIntegration : Velero Restore 통합 테스트
func TestVeleroRestoreIntegration(t *testing.T) {
	// Mock 클라이언트 설정
	mockClient := mocks.NewMockClient()

	// BaseHandler 생성 (Mock 사용)
	baseHandler := handler.NewBaseHandlerWithMock(nil)
	baseHandler.UseMockClient(true)

	// Velero Handler 생성
	veleroHandler := NewHandler(baseHandler)

	// Echo 컨텍스트 생성
	e := echo.New()
	req := e.NewRequest(nil, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := e.NewResponse(nil, nil)
	c := e.NewContext(req, rec)

	// JSON 바인딩을 위한 요청 본문 설정
	c.SetRequest(req)

	// 테스트 실행
	err := veleroHandler.GetRestores(c)

	// 결과 검증
	assert.NoError(t, err)
}

// TestVeleroBackupStorageLocationIntegration : Velero BackupStorageLocation 통합 테스트
func TestVeleroBackupStorageLocationIntegration(t *testing.T) {
	// Mock 클라이언트 설정
	mockClient := mocks.NewMockClient()

	// BaseHandler 생성 (Mock 사용)
	baseHandler := handler.NewBaseHandlerWithMock(nil)
	baseHandler.UseMockClient(true)

	// Velero Handler 생성
	veleroHandler := NewHandler(baseHandler)

	// Echo 컨텍스트 생성
	e := echo.New()
	req := e.NewRequest(nil, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := e.NewResponse(nil, nil)
	c := e.NewContext(req, rec)

	// JSON 바인딩을 위한 요청 본문 설정
	c.SetRequest(req)

	// 테스트 실행
	err := veleroHandler.GetBackupStorageLocations(c)

	// 결과 검증
	assert.NoError(t, err)
}

// TestVeleroVolumeSnapshotLocationIntegration : Velero VolumeSnapshotLocation 통합 테스트
func TestVeleroVolumeSnapshotLocationIntegration(t *testing.T) {
	// Mock 클라이언트 설정
	mockClient := mocks.NewMockClient()

	// BaseHandler 생성 (Mock 사용)
	baseHandler := handler.NewBaseHandlerWithMock(nil)
	baseHandler.UseMockClient(true)

	// Velero Handler 생성
	veleroHandler := NewHandler(baseHandler)

	// Echo 컨텍스트 생성
	e := echo.New()
	req := e.NewRequest(nil, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := e.NewResponse(nil, nil)
	c := e.NewContext(req, rec)

	// JSON 바인딩을 위한 요청 본문 설정
	c.SetRequest(req)

	// 테스트 실행
	err := veleroHandler.GetVolumeSnapshotLocations(c)

	// 결과 검증
	assert.NoError(t, err)
}

// TestVeleroBackupRepositoryIntegration : Velero BackupRepository 통합 테스트
func TestVeleroBackupRepositoryIntegration(t *testing.T) {
	// Mock 클라이언트 설정
	mockClient := mocks.NewMockClient()

	// BaseHandler 생성 (Mock 사용)
	baseHandler := handler.NewBaseHandlerWithMock(nil)
	baseHandler.UseMockClient(true)

	// Velero Handler 생성
	veleroHandler := NewHandler(baseHandler)

	// Echo 컨텍스트 생성
	e := echo.New()
	req := e.NewRequest(nil, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := e.NewResponse(nil, nil)
	c := e.NewContext(req, rec)

	// JSON 바인딩을 위한 요청 본문 설정
	c.SetRequest(req)

	// 테스트 실행
	err := veleroHandler.GetBackupRepositories(c)

	// 결과 검증
	assert.NoError(t, err)
}

// TestVeleroPodVolumeRestoreIntegration : Velero PodVolumeRestore 통합 테스트
func TestVeleroPodVolumeRestoreIntegration(t *testing.T) {
	// Mock 클라이언트 설정
	mockClient := mocks.NewMockClient()

	// BaseHandler 생성 (Mock 사용)
	baseHandler := handler.NewBaseHandlerWithMock(nil)
	baseHandler.UseMockClient(true)

	// Velero Handler 생성
	veleroHandler := NewHandler(baseHandler)

	// Echo 컨텍스트 생성
	e := echo.New()
	req := e.NewRequest(nil, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := e.NewResponse(nil, nil)
	c := e.NewContext(req, rec)

	// JSON 바인딩을 위한 요청 본문 설정
	c.SetRequest(req)

	// 테스트 실행
	err := veleroHandler.GetPodVolumeRestores(c)

	// 결과 검증
	assert.NoError(t, err)
}
