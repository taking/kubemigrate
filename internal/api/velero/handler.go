package velero

import (
	"context"
	"sort"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/api"
	"github.com/taking/kubemigrate/internal/errors"
	"github.com/taking/kubemigrate/internal/utils"
	"github.com/taking/kubemigrate/pkg/client"
)

// Handler : Velero 관련 HTTP 핸들러
type Handler struct {
	*api.BaseHandler
}

// NewHandler : 새로운 Velero 핸들러 생성
func NewHandler(base *api.BaseHandler) *Handler {
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
// @Failure 500 {object} errors.ErrorResponse
// @Router /v1/velero/health [post]
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.HandleResourceClient(c, "velero-health", func(client client.Client, ctx context.Context) (interface{}, error) {
		// Velero 연결 테스트
		_, err := client.Velero().GetBackups(ctx, "velero")
		if err != nil {
			return nil, errors.NewExternalError("velero", "GetBackups", err)
		}

		healthData := map[string]interface{}{
			"status": "UP",
		}
		return healthData, nil
	})
}

// === 백업 관련 핸들러 ===

// GetBackups : Velero 백업 목록 조회
// @Summary Get Backups
// @Description Get list of Velero backups
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Param namespace query string false "Namespace name (default: 'velero', all namespaces: 'all')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /v1/velero/backups [post]
func (h *Handler) GetBackups(c echo.Context) error {
	return h.HandleResourceClient(c, "backups", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "velero")

		// 백업 목록 조회
		backups, err := client.Velero().GetBackups(ctx, namespace)
		if err != nil {
			return nil, err
		}

		// 생성 시간 기준으로 정렬
		sort.Slice(backups, func(i, j int) bool {
			return backups[j].CreationTimestamp.Before(&backups[i].CreationTimestamp)
		})

		return backups, nil
	})
}

// GetBackupRepositories : 백업 저장소 목록 조회
// @Summary Get Backup Repositories
// @Description Get list of Velero backup repositories
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Param namespace query string false "Namespace name (default: 'velero', all namespaces: 'all')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /v1/velero/repositories [get]
func (h *Handler) GetBackupRepositories(c echo.Context) error {
	return h.HandleResourceClient(c, "backup-repositories", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "velero")

		// 백업 저장소 목록 조회
		repositories, err := client.Velero().GetBackupRepositories(ctx, namespace)
		if err != nil {
			return nil, err
		}

		return repositories, nil
	})
}

// GetBackupStorageLocations : 백업 스토리지 위치 목록 조회
// @Summary Get Backup Storage Locations
// @Description Get list of Velero backup storage locations
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Param namespace query string false "Namespace name (default: 'velero', all namespaces: 'all')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /v1/velero/storage-locations [get]
func (h *Handler) GetBackupStorageLocations(c echo.Context) error {
	return h.HandleResourceClient(c, "backup-storage-locations", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "velero")

		// 백업 스토리지 위치 목록 조회
		locations, err := client.Velero().GetBackupStorageLocations(ctx, namespace)
		if err != nil {
			return nil, err
		}

		return locations, nil
	})
}

// === 복구 관련 핸들러 ===

// GetRestores : Velero 복구 목록 조회
// @Summary Get Restores
// @Description Get list of Velero restores
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Param namespace query string false "Namespace name (default: 'velero', all namespaces: 'all')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /v1/velero/restores [post]
func (h *Handler) GetRestores(c echo.Context) error {
	return h.HandleResourceClient(c, "restores", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "velero")

		// 복구 목록 조회
		restores, err := client.Velero().GetRestores(ctx, namespace)
		if err != nil {
			return nil, err
		}

		// 생성 시간 기준으로 정렬
		sort.Slice(restores, func(i, j int) bool {
			return restores[j].CreationTimestamp.Before(&restores[i].CreationTimestamp)
		})

		return restores, nil
	})
}

// GetVolumeSnapshotLocations : 볼륨 스냅샷 위치 목록 조회
// @Summary Get Volume Snapshot Locations
// @Description Get list of Velero volume snapshot locations
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Param namespace query string false "Namespace name (default: 'velero', all namespaces: 'all')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /v1/velero/volume-snapshot-locations [get]
func (h *Handler) GetVolumeSnapshotLocations(c echo.Context) error {
	return h.HandleResourceClient(c, "volume-snapshot-locations", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "velero")

		// 볼륨 스냅샷 위치 목록 조회
		locations, err := client.Velero().GetVolumeSnapshotLocations(ctx, namespace)
		if err != nil {
			return nil, err
		}

		return locations, nil
	})
}

// GetPodVolumeRestores : Pod 볼륨 복구 목록 조회
// @Summary Get Pod Volume Restores
// @Description Get list of Velero pod volume restores
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Param namespace query string false "Namespace name (default: 'velero', all namespaces: 'all')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /v1/velero/pod-volume-restores [get]
func (h *Handler) GetPodVolumeRestores(c echo.Context) error {
	return h.HandleResourceClient(c, "pod-volume-restores", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "velero")

		// Pod 볼륨 복구 목록 조회
		restores, err := client.Velero().GetPodVolumeRestores(ctx, namespace)
		if err != nil {
			return nil, err
		}

		return restores, nil
	})
}
