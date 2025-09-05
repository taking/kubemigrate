package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/pkg/cache"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/health"
	"github.com/taking/kubemigrate/pkg/interfaces"
	_ "github.com/taking/kubemigrate/pkg/models"
	"github.com/taking/kubemigrate/pkg/response"
	"github.com/taking/kubemigrate/pkg/utils"
	"github.com/taking/kubemigrate/pkg/validator"
)

// VeleroHandler : Velero 관련 HTTP 요청을 처리하는 핸들러
type VeleroHandler struct {
	kubernetesValidator *validator.KubernetesValidator
	minioValidator      *validator.MinioValidator
	cache               *cache.Cache
	workerPool          *utils.WorkerPool
	healthManager       *health.HealthManager
}

// NewVeleroHandler : 새로운 VeleroHandler 인스턴스 생성
func NewVeleroHandler(appCache *cache.Cache, workerPool *utils.WorkerPool, healthManager *health.HealthManager) *VeleroHandler {
	return &VeleroHandler{
		kubernetesValidator: validator.NewKubernetesValidator(),
		minioValidator:      validator.NewMinioValidator(),
		cache:               appCache,
		workerPool:          workerPool,
		healthManager:       healthManager,
	}
}

// HealthCheck : Velero 연결 상태 확인
// @Summary Velero 연결 상태 확인
// @Description Validate Velero connection using KubeConfig
// @Tags velero
// @Accept json
// @Produce json
// @Param request body models.KubeConfig true "Velero connection configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Connection successful"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Failure 503 {object} models.SwaggerErrorResponse "Service unavailable"
// @Router /velero/health [get]
func (h *VeleroHandler) HealthCheck(c echo.Context) error {
	req, err := utils.BindAndValidateKubeConfig(c, h.kubernetesValidator)
	if err != nil {
		return err
	}

	// 기본 네임스페이스 설정
	req.Namespace = utils.ResolveNamespace(&req, c, "velero")

	client, err := client.NewVeleroClient(req)
	if err != nil {
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}

	// 클러스터 연결 상태 확인
	if err := client.HealthCheck(c.Request().Context()); err != nil {
		return response.RespondError(c, http.StatusServiceUnavailable,
			"Velero cluster unhealthy: "+err.Error())
	}

	// 헬스체크 매니저에 Velero 체커 등록
	if h.healthManager != nil {
		veleroChecker := health.NewVeleroHealthChecker(req)
		h.healthManager.Register(veleroChecker)
	}

	return response.RespondStatus(c, "healthy", "Velero connection successful")
}

// GetBackups : Velero 백업 목록 조회
// @Summary Velero 백업 목록 조회
// @Description Retrieve Velero backup list using MinioConfig and KubeConfig
// @Tags velero
// @Accept json
// @Produce json
// @Param request body models.VeleroConfig true "Velero connection configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Success"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /velero/backups [post]
func (h *VeleroHandler) GetBackups(c echo.Context) error {
	return h.handleVeleroResourceWithCache(c, "backups", func(client interfaces.VeleroClient, ctx context.Context) (interface{}, error) {
		data, err := client.GetBackups(ctx)
		if err != nil {
			return nil, err
		}

		// 생성 시간 순서대로 정렬
		sort.Slice(data, func(i, j int) bool {
			return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time) //nolint:staticcheck
		})

		return data, nil
	})
}

// GetRestores : Velero 복구 목록 조회
// @Summary Velero 복구 목록 조회
// @Description Retrieve Velero restore list using MinioConfig and KubeConfig
// @Tags velero
// @Accept json
// @Produce json
// @Param request body models.VeleroConfig true "Velero connection configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Success"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /velero/restores [get]
func (h *VeleroHandler) GetRestores(c echo.Context) error {
	return h.handleVeleroResourceWithCache(c, "restores", func(client interfaces.VeleroClient, ctx context.Context) (interface{}, error) {
		data, err := client.GetRestores(ctx)
		if err != nil {
			return nil, err
		}

		// 생성 시간 순서대로 정렬
		sort.Slice(data, func(i, j int) bool {
			return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time) //nolint:staticcheck
		})

		return data, nil
	})
}

// GetBackupRepositories : Velero 백업 저장소 목록 조회
// @Summary Velero 백업 저장소 목록 조회
// @Description Retrieve Velero backup repository list using MinioConfig and KubeConfig
// @Tags velero
// @Accept json
// @Produce json
// @Param request body models.VeleroConfig true "Velero connection configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Success"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /velero/backup-repositories [get]
func (h *VeleroHandler) GetBackupRepositories(c echo.Context) error {
	return h.handleVeleroResourceWithCache(c, "backup-repositories", func(client interfaces.VeleroClient, ctx context.Context) (interface{}, error) {
		data, err := client.GetBackupRepositories(ctx)
		if err != nil {
			return nil, err
		}

		// 생성 시간 순서대로 정렬
		sort.Slice(data, func(i, j int) bool {
			return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time) //nolint:staticcheck
		})

		return data, nil
	})
}

// GetBackupStorageLocations : Velero 백업 저장 위치 목록 조회
// @Summary Velero 백업 저장 위치 목록 조회
// @Description Retrieve Velero backup storage location list using MinioConfig and KubeConfig
// @Tags velero
// @Accept json
// @Produce json
// @Param request body models.VeleroConfig true "Velero connection configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Success"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /velero/backup-storage-locations [get]
func (h *VeleroHandler) GetBackupStorageLocations(c echo.Context) error {
	return h.handleVeleroResourceWithCache(c, "backup-storage-locations", func(client interfaces.VeleroClient, ctx context.Context) (interface{}, error) {
		data, err := client.GetBackupStorageLocations(ctx)
		if err != nil {
			return nil, err
		}

		// 생성 시간 순서대로 정렬
		sort.Slice(data, func(i, j int) bool {
			return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time) //nolint:staticcheck
		})

		return data, nil
	})
}

// GetVolumeSnapshotLocations : Velero 볼륨 스냅샷 위치 목록 조회
// @Summary Velero 볼륨 스냅샷 위치 목록 조회
// @Description Retrieve Velero volume snapshot location list using MinioConfig and KubeConfig
// @Tags velero
// @Accept json
// @Produce json
// @Param request body models.VeleroConfig true "Velero connection configuration"
// @Success 200 {object} models.SwaggerSuccessResponse "Success"
// @Failure 400 {object} models.SwaggerErrorResponse "Bad request"
// @Failure 500 {object} models.SwaggerErrorResponse "Internal server error"
// @Router /velero/volume-snapshot-locations [get]
func (h *VeleroHandler) GetVolumeSnapshotLocations(c echo.Context) error {
	return h.handleVeleroResourceWithCache(c, "volume-snapshot-locations", func(client interfaces.VeleroClient, ctx context.Context) (interface{}, error) {
		data, err := client.GetVolumeSnapshotLocations(ctx)
		if err != nil {
			return nil, err
		}

		// 생성 시간 순서대로 정렬
		sort.Slice(data, func(i, j int) bool {
			return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time) //nolint:staticcheck
		})

		return data, nil
	})
}

// handleVeleroResourceWithCache : 캐시를 사용하는 Velero 리소스 처리 헬퍼
func (h *VeleroHandler) handleVeleroResourceWithCache(c echo.Context, cacheKey string,
	getResource func(interfaces.VeleroClient, context.Context) (interface{}, error)) error {

	// VeleroConfig 검증
	req, err := utils.BindAndValidateVeleroConfig(c, h.minioValidator, h.kubernetesValidator)
	if err != nil {
		return err
	}

	// 기본 네임스페이스 설정
	req.Namespace = utils.ResolveNamespace(&req.KubeConfig, c, "velero")

	// 캐시 키 생성
	fullCacheKey := fmt.Sprintf("%s:%s", cacheKey, req.Namespace)

	// 캐시에서 가져오기
	if cached, exists := h.cache.Get(fullCacheKey); exists {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   cached,
			"cached": true,
		})
	}

	// 캐시에 없으면 클라이언트에서 가져오기
	veleroClient, err := client.NewVeleroClient(req.KubeConfig)
	if err != nil {
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}

	// 워커 풀을 사용하여 백그라운드에서 데이터 가져오기
	resultChan := make(chan interface{}, 1)
	errorChan := make(chan error, 1)

	h.workerPool.Submit(func() {
		data, err := getResource(veleroClient, c.Request().Context())
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- data
	})

	select {
	case data := <-resultChan:
		// 캐시에 저장
		h.cache.Set(fullCacheKey, data)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data":   data,
			"cached": false,
		})
	case err := <-errorChan:
		return response.RespondError(c, http.StatusInternalServerError, err.Error())
	}
}
