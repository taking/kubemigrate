package usecase

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/taking/velero/internal/repository"
	"github.com/taking/velero/pkg/response"

	"github.com/taking/velero/internal/client"
	"net/http"
	"sort"
)

// VeleroController : Velero 관련 API 컨트롤러
type VeleroController struct {
	*BaseController
}

func NewVeleroController() *VeleroController {
	return &VeleroController{
		BaseController: NewBaseController(),
	}
}

// CheckVeleroConnection : Kubernetes 클러스터 Velero 연결 확인
// CheckVeleroConnection godoc
// @Summary Velero 연결 확인
// @Description KubeConfig을 사용하여 Velero 연결 검증
// @Tags velero
// @Accept json
// @Produce json
// @Param request body github.com/taking/velero/internal/model.KubeConfig true "Velero 연결에 필요한 값"
// @Success 200 {object} model.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} model.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} model.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} model.SwaggerErrorResponse "서비스 이용 불가"
// @Router /velero/health [get]
func (c *VeleroController) CheckVeleroConnection(ctx echo.Context) error {
	req, err := c.BindAndValidateKubeConfig(ctx)
	if err != nil {
		return err
	}

	// 네임스페이스 값이 없으면, 기본 네임스페이스 "velero"로 설정
	req.Namespace = c.ResolveNamespace(&req, ctx, "velero")

	client, err := client.NewVeleroClient(req)
	if err != nil {
		return response.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	// 클러스터 연결 상태 확인
	return c.HandleHealthCheck(ctx, client, "Velero")
}

// handleVeleroResource : 공통 리소스 처리 헬퍼
// 코드 중복을 줄이기 위해 Velero 리소스 조회 후 JSON 반환을 공통화
func (c *VeleroController) handleVeleroResource(ctx echo.Context,
	getResource func(repository.VeleroClient, context.Context) (interface{}, error)) error {

	// VeleroConfig 바인딩 및 검증
	req, err := c.BindAndValidateVeleroConfig(ctx)
	if err != nil {
		return err
	}

	// 네임스페이스 값이 없으면, 기본 네임스페이스 "velero"로 설정
	req.Namespace = c.ResolveNamespace(&req.KubeConfig, ctx, "velero")

	veleroClient, err := client.NewVeleroClient(req.KubeConfig)
	if err != nil {
		return response.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	data, err := getResource(veleroClient, context.Background())
	if err != nil {
		return response.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	// 데이터가 정렬 가능한 경우, 생성 시간 기준으로 정렬
	if sortable, ok := data.(interface {
		Len() int
		Swap(i, j int)
		Less(i, j int) bool
	}); ok {
		sort.Sort(sortable)
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

// GetBackups : Velero 백업 목록 조회
// GetBackups godoc
// @Summary Velero 백업 목록 확인
// @Description MinioConfig, KubeConfig을 사용하여 Velero 백업 목록 조회
// @Tags velero
// @Accept json
// @Produce json
// @Param request body github.com/taking/velero/internal/model.VeleroConfigRequest true "Velero 연결에 필요한 값"
// @Success 200 {object} model.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} model.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} model.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} model.SwaggerErrorResponse "서비스 이용 불가"
// @Router /velero/backups [post]
func (c *VeleroController) GetBackups(ctx echo.Context) error {
	return c.handleVeleroResource(ctx, func(client repository.VeleroClient, ctx context.Context) (interface{}, error) {
		data, err := client.GetBackups(ctx)
		if err != nil {
			return nil, err
		}

		// 최신 생성일 순으로 정렬
		sort.Slice(data, func(i, j int) bool {
			return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
		})

		return data, nil
	})
}

// GetRestores : Velero 복구 목록 조회
// GetRestores godoc
// @Summary Velero 복구 목록 확인
// @Description MinioConfig, KubeConfig을 사용하여 Velero 복구 목록 조회
// @Tags velero
// @Accept json
// @Produce json
// @Param request body github.com/taking/velero/internal/model.VeleroConfig true "Velero 연결에 필요한 값"
// @Success 200 {object} model.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} model.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} model.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} model.SwaggerErrorResponse "서비스 이용 불가"
// @Router /velero/restores [get]
func (c *VeleroController) GetRestores(ctx echo.Context) error {
	return c.handleVeleroResource(ctx, func(client repository.VeleroClient, ctx context.Context) (interface{}, error) {
		data, err := client.GetRestores(ctx)
		if err != nil {
			return nil, err
		}

		// 최신 생성일 순으로 정렬
		sort.Slice(data, func(i, j int) bool {
			return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
		})

		return data, nil
	})
}

// GetBackupRepositories : 백업 저장소 목록 조회
// GetBackupRepositories godoc
// @Summary Velero 백업 저장소 목록 확인
// @Description MinioConfig, KubeConfig을 사용하여 Velero 백업 저장소 목록 조회
// @Tags velero
// @Accept json
// @Produce json
// @Param request body github.com/taking/velero/internal/model.VeleroConfig true "Velero 연결에 필요한 값"
// @Success 200 {object} model.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} model.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} model.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} model.SwaggerErrorResponse "서비스 이용 불가"
// @Router /velero/backup-repositories [get]
func (c *VeleroController) GetBackupRepositories(ctx echo.Context) error {
	return c.handleVeleroResource(ctx, func(client repository.VeleroClient, ctx context.Context) (interface{}, error) {
		data, err := client.GetBackupRepositories(ctx)
		if err != nil {
			return nil, err
		}

		// 최신 생성일 순으로 정렬
		sort.Slice(data, func(i, j int) bool {
			return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
		})

		return data, nil
	})
}

// GetBackupStorageLocations : 백업 스토리지 위치 목록 조회
// GetBackupStorageLocations godoc
// @Summary Velero 백업 스토리지 위치 목록 확인
// @Description MinioConfig, KubeConfig을 사용하여 Velero 백업 스토리지 위치 목록 조회
// @Tags velero
// @Accept json
// @Produce json
// @Param request body github.com/taking/velero/internal/model.VeleroConfig true "Velero 연결에 필요한 값"
// @Success 200 {object} model.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} model.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} model.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} model.SwaggerErrorResponse "서비스 이용 불가"
// @Router /velero/backup-storage-locations [get]
func (c *VeleroController) GetBackupStorageLocations(ctx echo.Context) error {
	return c.handleVeleroResource(ctx, func(client repository.VeleroClient, ctx context.Context) (interface{}, error) {
		data, err := client.GetBackupStorageLocations(ctx)
		if err != nil {
			return nil, err
		}

		// 최신 생성일 순으로 정렬
		sort.Slice(data, func(i, j int) bool {
			return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
		})

		return data, nil
	})
}

// GetVolumeSnapshotLocations : 볼륨 스냅샷 위치 목록 조회
// GetVolumeSnapshotLocations godoc
// @Summary Velero 볼륨 스냅샷 위치 목록 확인
// @Description MinioConfig, KubeConfig을 사용하여 Velero 볼륨 스냅샷 위치 목록 조회
// @Tags velero
// @Accept json
// @Produce json
// @Param request body github.com/taking/velero/internal/model.VeleroConfig true "Velero 연결에 필요한 값"
// @Success 200 {object} model.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} model.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} model.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} model.SwaggerErrorResponse "서비스 이용 불가"
// @Router /velero/volume-snapshot-locations [get]
func (c *VeleroController) GetVolumeSnapshotLocations(ctx echo.Context) error {
	return c.handleVeleroResource(ctx, func(client repository.VeleroClient, ctx context.Context) (interface{}, error) {
		data, err := client.GetVolumeSnapshotLocations(ctx)
		if err != nil {
			return nil, err
		}

		// 최신 생성일 순으로 정렬
		sort.Slice(data, func(i, j int) bool {
			return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
		})

		return data, nil
	})
}

// generateBackupSummary: Backup 요약 생성을 위한 헬퍼 메서드
//func generateBackupSummary(backups []velerov1.Backup) model.BackupSummary {
//	summary := model.BackupSummary{Total: len(backups)}
//	for _, b := range backups {
//		switch b.Status.Phase {
//		case velerov1.BackupPhaseCompleted:
//			summary.Completed++
//		case velerov1.BackupPhaseFailed:
//			summary.Failed++
//		case velerov1.BackupPhaseInProgress:
//			summary.InProgress++
//		case velerov1.BackupPhasePartiallyFailed:
//			summary.PartiallyFailed++
//		}
//		if time.Since(b.CreationTimestamp.Time) < 24*time.Hour {
//			summary.Recent++
//		}
//		if b.Status.Expiration != nil && b.Status.Expiration.Time.Before(time.Now()) {
//			summary.Expired++
//		}
//	}
//	return summary
//}

//// CompareStorageClasses: 스토리지 클래스 비교
//func (c *VeleroController) CompareStorageClasses(ctx echo.Context) error {
//	var req model.VeleroRequest
//	if err := ctx.Bind(&req); err != nil {
//		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
//	}
//
//	if req.SourceKubeConfig == "" || req.DestinationKubeConfig == "" {
//		return echo.NewHTTPError(http.StatusBadRequest, "Both source and destination kubeconfigs are required")
//	}
//
//	SourceKubeConfig, _ := pkg.DecodeIfBase64(req.SourceKubeConfig)
//
//	// 소스와 대상 클러스터의 스토리지 클래스를 가져오는 로직
//	// 여기서는 간단한 예시로 구현
//	sourceService, err := client.NewVeleroClientFromRawConfig(SourceKubeConfig)
//	if err != nil {
//		return echo.NewHTTPError(http.StatusBadRequest, "Invalid source kubeconfig: "+err.Error())
//	}
//
//	DestinationKubeConfig, _ := pkg.DecodeIfBase64(req.DestinationKubeConfig)
//
//	destService, err := client.NewVeleroClientFromRawConfig(DestinationKubeConfig)
//	if err != nil {
//		return echo.NewHTTPError(http.StatusBadRequest, "Invalid destination kubeconfig: "+err.Error())
//	}
//
//	namespace := req.Namespace
//	if namespace == "" {
//		namespace = "velero"
//	}
//
//	sourceLocations, err := sourceService.GetBackupStorageLocations(ctx.Request().Context(), namespace)
//	if err != nil {
//		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get source storage locations: "+err.Error())
//	}
//
//	destLocations, err := destService.GetBackupStorageLocations(ctx.Request().Context(), namespace)
//	if err != nil {
//		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get destination storage locations: "+err.Error())
//	}
//
//	return ctx.JSON(http.StatusOK, map[string]interface{}{
//		"sourceStorageLocations":      len(sourceLocations),
//		"destinationStorageLocations": len(destLocations),
//		"compatible":                  len(sourceLocations) > 0 && len(destLocations) > 0,
//		"message":                     "Storage location comparison completed",
//	})
//}
