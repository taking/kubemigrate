package controller

import (
	"context"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	//"fmt"
	"github.com/labstack/echo/v4"

	"net/http"
	"sort"
	"taking.kr/velero/clients"
	"taking.kr/velero/models"
	//"taking.kr/velero/utils"
	"taking.kr/velero/validation"
	"time"
)

type VeleroController struct {
	validator *validation.RequestValidator
}

func NewVeleroController() *VeleroController {
	return &VeleroController{
		validator: validation.NewRequestValidator(),
	}
}

// HealthCheck : Kubernetes 클러스터 Velero 연결 확인
func (c *VeleroController) HealthCheck(ctx echo.Context) error {
	var req models.KubeConfig

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// namespace 설정
	namespace := c.determineNamespace(&req, ctx)

	// kubeconfig 유효성 검사
	decodeKubeConfig, err := c.validator.ValidateKubeConfigRequest(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Velero 클라이언트 생성
	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: decodeKubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"Invalid kubeconfig: ": err.Error(),
		})
	}

	if err := client.HealthCheck(context.Background()); err != nil {
		return ctx.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"status":  "healthy",
		"message": "Kubernetes connection successful",
	})
}

func (c *VeleroController) determineNamespace(req *models.KubeConfig, ctx echo.Context) string {
	if req.Namespace != "" {
		return req.Namespace
	}
	if ns := ctx.QueryParam("namespace"); ns != "" {
		return ns
	}
	return "velero" // default
}

func (c *VeleroController) GetBackups(ctx echo.Context) error {
	var req models.KubeConfig

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// namespace 설정
	namespace := c.determineNamespace(&req, ctx)

	// kubeconfig 유효성 검사
	decodeKubeConfig, err := c.validator.ValidateKubeConfigRequest(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Velero 클라이언트 생성
	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: decodeKubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"Invalid kubeconfig: ": err.Error(),
		})
	}

	data, err := client.GetBackups(context.Background())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Sort backups by creation time (newest first)
	sort.Slice(data, func(i, j int) bool {
		return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
	})

	// Add summary information
	summary := c.generateBackupSummary(data)

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"data":    data,
		"summary": summary,
	})
}

func (c *VeleroController) GetRestores(ctx echo.Context) error {
	var req models.KubeConfig

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// namespace 설정
	namespace := c.determineNamespace(&req, ctx)

	// kubeconfig 유효성 검사
	decodeKubeConfig, err := c.validator.ValidateKubeConfigRequest(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Velero 클라이언트 생성
	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: decodeKubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"Invalid kubeconfig: ": err.Error(),
		})
	}

	data, err := client.GetRestores(context.Background())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Sort restores by creation time (newest first)
	sort.Slice(data, func(i, j int) bool {
		return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
	})

	// Add summary information
	summary := c.generateRestoreSummary(data)

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"data":    data,
		"summary": summary,
	})
}

func (c *VeleroController) GetBackupRepositories(ctx echo.Context) error {
	var req models.KubeConfig

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// namespace 설정
	namespace := c.determineNamespace(&req, ctx)

	// kubeconfig 유효성 검사
	decodeKubeConfig, err := c.validator.ValidateKubeConfigRequest(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Velero 클라이언트 생성
	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: decodeKubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"Invalid kubeconfig: ": err.Error(),
		})
	}

	data, err := client.GetBackupRepositories(context.Background())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Sort backupRepositories by creation time (newest first)
	sort.Slice(data, func(i, j int) bool {
		return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
	})

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

func (c *VeleroController) GetBackupStorageLocations(ctx echo.Context) error {
	var req models.KubeConfig

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// namespace 설정
	namespace := c.determineNamespace(&req, ctx)

	// kubeconfig 유효성 검사
	decodeKubeConfig, err := c.validator.ValidateKubeConfigRequest(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Velero 클라이언트 생성
	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: decodeKubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"Invalid kubeconfig: ": err.Error(),
		})
	}

	data, err := client.GetBackupStorageLocations(context.Background())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Sort backupStorageLocations by creation time (newest first)
	sort.Slice(data, func(i, j int) bool {
		return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
	})

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

func (c *VeleroController) GetVolumeSnapshotLocations(ctx echo.Context) error {
	var req models.KubeConfig

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// namespace 설정
	namespace := c.determineNamespace(&req, ctx)

	// kubeconfig 유효성 검사
	decodeKubeConfig, err := c.validator.ValidateKubeConfigRequest(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Velero 클라이언트 생성
	client, err := clients.NewVeleroClient(models.KubeConfig{
		KubeConfig: decodeKubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"Invalid kubeconfig: ": err.Error(),
		})
	}

	data, err := client.GetVolumeSnapshotLocations(context.Background())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Sort volumeSnapshotLocations by creation time (newest first)
	sort.Slice(data, func(i, j int) bool {
		return data[i].CreationTimestamp.Time.After(data[j].CreationTimestamp.Time)
	})

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

// Helper methods for summary generation
func (c *VeleroController) generateBackupSummary(backups []velerov1.Backup) models.BackupSummary {
	summary := models.BackupSummary{
		Total: len(backups),
	}

	for _, backup := range backups {
		switch backup.Status.Phase {
		case velerov1.BackupPhaseCompleted:
			summary.Completed++
		case velerov1.BackupPhaseFailed:
			summary.Failed++
		case velerov1.BackupPhaseInProgress:
			summary.InProgress++
		case velerov1.BackupPhasePartiallyFailed:
			summary.PartiallyFailed++
		}

		// Check for recent backups (last 24 hours)
		if time.Since(backup.CreationTimestamp.Time) < 24*time.Hour {
			summary.Recent++
		}

		// Check for expired backups
		if backup.Status.Expiration != nil && backup.Status.Expiration.Time.Before(time.Now()) {
			summary.Expired++
		}
	}

	return summary
}

func (c *VeleroController) generateRestoreSummary(restores []velerov1.Restore) models.RestoreSummary {
	summary := models.RestoreSummary{
		Total: len(restores),
	}

	for _, restore := range restores {
		switch restore.Status.Phase {
		case velerov1.RestorePhaseCompleted:
			summary.Completed++
		case velerov1.RestorePhaseFailed:
			summary.Failed++
		case velerov1.RestorePhaseInProgress:
			summary.InProgress++
		case velerov1.RestorePhasePartiallyFailed:
			summary.PartiallyFailed++
		}

		// Check for recent restores (last 24 hours)
		if time.Since(restore.CreationTimestamp.Time) < 24*time.Hour {
			summary.Recent++
		}
	}
	return summary
}

//func (c *VeleroController) generateStorageLocationSummary(locations []velerov1.BackupStorageLocation) models.StorageLocationSummary {
//	summary := models.StorageLocationSummary{
//		Total: len(locations),
//	}
//
//	for _, location := range locations {
//		switch location.Status.Phase {
//		case velerov1.BackupStorageLocationPhaseAvailable:
//			summary.Available++
//		case velerov1.BackupStorageLocationPhaseUnavailable:
//			summary.Unavailable++
//		}
//	}
//
//	return summary
//}
//
//func (c *VeleroController) formatBytes(bytes int64) string {
//	const unit = 1024
//	if bytes < unit {
//		return fmt.Sprintf("%d B", bytes)
//	}
//	div, exp := int64(unit), 0
//	for n := bytes / unit; n >= unit; n /= unit {
//		div *= unit
//		exp++
//	}
//	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
//}

//// 스토리지 클래스 비교
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
