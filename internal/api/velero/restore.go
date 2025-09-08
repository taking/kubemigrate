package velero

import (
	"context"
	"sort"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/pkg/client/velero"
	"github.com/taking/kubemigrate/pkg/utils"
)

// GetRestores : Velero 복구 목록 조회
// @Summary Get Restores
// @Description Get list of Velero restores
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/velero/restores [post]
func (h *Handler) GetRestores(c echo.Context) error {
	return h.HandleVeleroResource(c, "restores", func(veleroClient velero.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		req, err := utils.BindAndValidateVeleroConfig(c, h.MinioValidator, h.KubernetesValidator)
		if err != nil {
			return nil, err
		}

		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(&req.KubeConfig, c, "velero")

		// 복구 목록 조회
		restores, err := veleroClient.GetRestores(ctx, namespace)
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
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/velero/volume-snapshot-locations [get]
func (h *Handler) GetVolumeSnapshotLocations(c echo.Context) error {
	return h.HandleVeleroResource(c, "volume-snapshot-locations", func(veleroClient velero.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		req, err := utils.BindAndValidateVeleroConfig(c, h.MinioValidator, h.KubernetesValidator)
		if err != nil {
			return nil, err
		}

		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(&req.KubeConfig, c, "velero")

		// 볼륨 스냅샷 위치 목록 조회
		locations, err := veleroClient.GetVolumeSnapshotLocations(ctx, namespace)
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
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/velero/pod-volume-restores [get]
func (h *Handler) GetPodVolumeRestores(c echo.Context) error {
	return h.HandleVeleroResource(c, "pod-volume-restores", func(veleroClient velero.Client, ctx context.Context) (interface{}, error) {
		// 요청 바인딩 및 검증
		req, err := utils.BindAndValidateVeleroConfig(c, h.MinioValidator, h.KubernetesValidator)
		if err != nil {
			return nil, err
		}

		// 네임스페이스 결정
		namespace := utils.ResolveNamespace(&req.KubeConfig, c, "velero")

		// Pod 볼륨 복구 목록 조회
		restores, err := veleroClient.GetPodVolumeRestores(ctx, namespace)
		if err != nil {
			return nil, err
		}

		return restores, nil
	})
}
