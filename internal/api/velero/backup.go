package velero

import (
	"context"
	"sort"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/pkg/client/velero"
	"github.com/taking/kubemigrate/pkg/utils"
)

// GetBackups : Velero 백업 목록 조회
// @Summary Get Backups
// @Description Get list of Velero backups
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Param namespace query string false "Namespace name (default: 'velero', all namespaces: 'all')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/backups [post]
func (h *Handler) GetBackups(c echo.Context) error {
	return h.HandleVeleroResource(c, "backups", func(veleroClient velero.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		_, err := utils.BindAndValidateVeleroConfig(c, h.MinioValidator, h.KubernetesValidator)
		if err != nil {
			return nil, err
		}

		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "velero")

		// 백업 목록 조회
		backups, err := veleroClient.GetBackups(ctx, namespace)
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
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/repositories [get]
func (h *Handler) GetBackupRepositories(c echo.Context) error {
	return h.HandleVeleroResource(c, "backup-repositories", func(veleroClient velero.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		_, err := utils.BindAndValidateVeleroConfig(c, h.MinioValidator, h.KubernetesValidator)
		if err != nil {
			return nil, err
		}

		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "velero")

		// 백업 저장소 목록 조회
		repositories, err := veleroClient.GetBackupRepositories(ctx, namespace)
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
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/storage-locations [get]
func (h *Handler) GetBackupStorageLocations(c echo.Context) error {
	return h.HandleVeleroResource(c, "backup-storage-locations", func(veleroClient velero.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		_, err := utils.BindAndValidateVeleroConfig(c, h.MinioValidator, h.KubernetesValidator)
		if err != nil {
			return nil, err
		}

		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(c, "velero")

		// 백업 스토리지 위치 목록 조회
		locations, err := veleroClient.GetBackupStorageLocations(ctx, namespace)
		if err != nil {
			return nil, err
		}

		return locations, nil
	})
}
