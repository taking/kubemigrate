package velero

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
)

// VeleroInstallConfig : Velero 설치 설정
type VeleroInstallConfig struct {
	MinioConfig config.MinioConfig `json:"minio_config"`
}

// InstallResult : 설치 결과
type InstallResult struct {
	Status           string                 `json:"status"`
	Message          string                 `json:"message"`
	VeleroNamespace  string                 `json:"velero_namespace"`
	MinioConnected   bool                   `json:"minio_connected"`
	BackupLocation   string                 `json:"backup_location"`
	InstallationTime time.Duration          `json:"installation_time"`
	Details          map[string]interface{} `json:"details,omitempty"`
}

// Handler : Velero 관련 HTTP 핸들러
type Handler struct {
	*handler.BaseHandler
}

// NewHandler : 새로운 Velero 핸들러 생성
func NewHandler(base *handler.BaseHandler) *Handler {
	return &Handler{
		BaseHandler: base,
	}
}

// HealthCheck : Velero 연결 테스트
// @Summary Velero Connection Test
// @Description Test Velero connection with provided configuration
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/health [post]
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.HandleResourceClient(c, "velero-health", func(client client.Client, ctx context.Context) (interface{}, error) {
		// Velero 연결 테스트
		_, err := client.Velero().GetBackups(ctx, "velero")
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"service": "velero",
			"message": "Velero connection is working",
		}, nil
	})
}

// InstallVeleroWithMinIO : Velero 설치 및 MinIO 연동 설정
// @Summary Install Velero with MinIO
// @Description Install Velero and configure MinIO integration
// @Tags velero
// @Accept json
// @Produce json
// @Param request body velero.VeleroInstallConfig true "Velero installation configuration (minio_config only)"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/install [post]
func (h *Handler) InstallVeleroWithMinIO(c echo.Context) error {
	// HandleResourceClient 패턴 사용
	return h.InstallVeleroWithMinIOHandler(c)
}

// InstallVeleroWithMinIOHandler : HandleResourceClient 패턴으로 Velero 설치 및 MinIO 연동
func (h *Handler) InstallVeleroWithMinIOHandler(c echo.Context) error {
	return h.HandleResourceClient(c, "velero-install", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩
		var config VeleroInstallConfig
		if err := c.Bind(&config); err != nil {
			return nil, fmt.Errorf("invalid request: %w", err)
		}

		// Velero 설치 및 MinIO 연동 실행
		return h.InstallVeleroWithMinIOInternal(client, ctx, config)
	})
}

// InstallVeleroWithMinIOInternal : Velero 설치 및 MinIO 연동 설정 (내부 로직)
func (h *Handler) InstallVeleroWithMinIOInternal(client client.Client, ctx context.Context, config VeleroInstallConfig) (*InstallResult, error) {
	startTime := time.Now()

	// 고정값 설정
	namespace := "velero"

	result := &InstallResult{
		Status:          "in_progress",
		VeleroNamespace: namespace,
		Details:         make(map[string]interface{}),
	}

	// 1. Velero 설치 여부 확인
	if err := h.checkVeleroInstallation(client, ctx, namespace); err != nil {
		return nil, fmt.Errorf("velero installation check failed: %w", err)
	}

	// 2. Velero가 없으면 Helm으로 설치
	if !h.isVeleroInstalled(client, ctx, namespace) {
		if err := h.installVeleroViaHelm(client, ctx, config); err != nil {
			return nil, fmt.Errorf("velero installation failed: %w", err)
		}
		result.Details["installation"] = "completed"
	} else {
		result.Details["installation"] = "already_installed"
	}

	// 3. Velero readiness 재확인
	if err := h.waitForVeleroReady(client, ctx, namespace); err != nil {
		return nil, fmt.Errorf("velero readiness check failed: %w", err)
	}
	result.Details["readiness"] = "ready"

	// 4. MinIO Secret 생성
	if err := h.createMinIOSecret(client, ctx, config, namespace); err != nil {
		return nil, fmt.Errorf("minio secret creation failed: %w", err)
	}
	result.Details["minio_secret"] = "created"

	// 5. BackupStorageLocation 생성
	_, err := h.createBackupStorageLocation(client, ctx, config, namespace)
	if err != nil {
		return nil, fmt.Errorf("backup storage location creation failed: %w", err)
	}
	result.Details["backup_location"] = "created"
	result.BackupLocation = fmt.Sprintf("minio://%s", config.MinioConfig.Endpoint)

	// 6. BSL 상태 조회 및 MinIO 연결 검증
	minioConnected, err := h.validateMinIOConnection(client, ctx, config, namespace)
	if err != nil {
		return nil, fmt.Errorf("minio connection validation failed: %w", err)
	}
	result.Details["minio_validation"] = "completed"
	result.MinioConnected = minioConnected

	result.Status = "success"
	result.Message = "Velero installed and configured successfully"
	result.InstallationTime = time.Since(startTime)

	return result, nil
}
