package controller

import (
	"context"
	"github.com/labstack/echo/v4"
	"taking.kr/velero/helpers"

	"net/http"
	"sort"
	"taking.kr/velero/clients"
	"taking.kr/velero/models"
	"taking.kr/velero/validation"
)

// VeleroController : Velero 관련 API 컨트롤러
type VeleroController struct {
	validator *validation.RequestValidator
}

func NewVeleroController() *VeleroController {
	return &VeleroController{
		validator: validation.NewRequestValidator(),
	}
}

// CheckVeleroConnection : Kubernetes 클러스터 Velero 연결 확인
func (c *VeleroController) CheckVeleroConnection(ctx echo.Context) error {
	req, err := helpers.BindAndValidateKubeConfig(ctx, c.validator)
	if err != nil {
		return err
	}

	namespace := helpers.ResolveNamespace(&req, ctx, "velero")

	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: req.KubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	if err := client.HealthCheck(ctx.Request().Context()); err != nil {
		return helpers.JSONError(ctx, http.StatusServiceUnavailable, "Velero cluster unhealthy: "+err.Error())
	}

	return helpers.JSONStatus(ctx, "healthy", "Kubernetes connection successful")
}

func (c *VeleroController) GetBackups(ctx echo.Context) error {
	req, err := helpers.BindAndValidateKubeConfig(ctx, c.validator)
	if err != nil {
		return err
	}
	namespace := helpers.ResolveNamespace(&req, ctx, "velero")

	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: req.KubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	data, err := client.GetBackups(context.Background())
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
	})

	//summary := generateBackupSummary(data)

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
		//"summary": summary,
	})
}

func (c *VeleroController) GetRestores(ctx echo.Context) error {
	req, err := helpers.BindAndValidateKubeConfig(ctx, c.validator)
	if err != nil {
		return err
	}
	namespace := helpers.ResolveNamespace(&req, ctx, "velero")

	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: req.KubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	data, err := client.GetRestores(context.Background())
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
	})

	//summary := generateRestoresSummary(data)

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
		//"summary": summary,
	})
}

func (c *VeleroController) GetBackupRepositories(ctx echo.Context) error {
	req, err := helpers.BindAndValidateKubeConfig(ctx, c.validator)
	if err != nil {
		return err
	}
	namespace := helpers.ResolveNamespace(&req, ctx, "velero")

	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: req.KubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	data, err := client.GetBackupRepositories(context.Background())
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
	})

	//summary := generateRestoresSummary(data)

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
		//"summary": summary,
	})
}

func (c *VeleroController) GetBackupStorageLocations(ctx echo.Context) error {
	req, err := helpers.BindAndValidateKubeConfig(ctx, c.validator)
	if err != nil {
		return err
	}
	namespace := helpers.ResolveNamespace(&req, ctx, "velero")

	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: req.KubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	data, err := client.GetBackupStorageLocations(context.Background())
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
	})

	//summary := generateRestoresSummary(data)

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
		//"summary": summary,
	})
}

func (c *VeleroController) GetVolumeSnapshotLocations(ctx echo.Context) error {
	req, err := helpers.BindAndValidateKubeConfig(ctx, c.validator)
	if err != nil {
		return err
	}
	namespace := helpers.ResolveNamespace(&req, ctx, "velero")

	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: req.KubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	data, err := client.GetVolumeSnapshotLocations(context.Background())
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
	})

	//summary := generateRestoresSummary(data)

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
		//"summary": summary,
	})
}

// generateBackupSummary: Backup 요약 생성을 위한 헬퍼 메서드
//func generateBackupSummary(backups []velerov1.Backup) models.BackupSummary {
//	summary := models.BackupSummary{Total: len(backups)}
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
//	var req models.VeleroRequest
//	if err := ctx.Bind(&req); err != nil {
//		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
//	}
//
//	if req.SourceKubeConfig == "" || req.DestinationKubeConfig == "" {
//		return echo.NewHTTPError(http.StatusBadRequest, "Both source and destination kubeconfigs are required")
//	}
//
//	SourceKubeConfig, _ := utils.DecodeIfBase64(req.SourceKubeConfig)
//
//	// 소스와 대상 클러스터의 스토리지 클래스를 가져오는 로직
//	// 여기서는 간단한 예시로 구현
//	sourceService, err := clients.NewVeleroClientFromRawConfig(SourceKubeConfig)
//	if err != nil {
//		return echo.NewHTTPError(http.StatusBadRequest, "Invalid source kubeconfig: "+err.Error())
//	}
//
//	DestinationKubeConfig, _ := utils.DecodeIfBase64(req.DestinationKubeConfig)
//
//	destService, err := clients.NewVeleroClientFromRawConfig(DestinationKubeConfig)
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
