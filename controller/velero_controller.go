package controller

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"log"
	"net/http"
	"sort"
	"taking.kr/velero/clients"
	"taking.kr/velero/interfaces"
	"taking.kr/velero/models"
	"taking.kr/velero/utils"
	"taking.kr/velero/validation"
	"time"
)

type VeleroBodyController struct {
	validator *validation.RequestValidator
}

func NewVeleroBodyController() *VeleroBodyController {
	return &VeleroBodyController{
		validator: validation.NewRequestValidator(),
	}
}

// getVeleroServiceFromRequest는 요청에서 kubeconfig를 추출하여 서비스를 생성합니다
func (c *VeleroBodyController) getVeleroServiceFromRequest(ctx echo.Context) (interfaces.VeleroService, string, error) {
	var req models.VeleroRequest

	if err := ctx.Bind(&req); err != nil {
		return nil, "", echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// kubeconfig 유효성 검사
	if req.SourceKubeconfig == "" {
		return nil, "", echo.NewHTTPError(http.StatusBadRequest, "sourceKubeconfig is required")
	}

	// Validate request
	if err := c.validator.ValidateVeleroRequest(&req); err != nil {
		return nil, "", echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// namespace 설정
	namespace := c.determineNamespace(&req, ctx)

	decodeKubeconfig, err := utils.DecodeIfBase64(req.SourceKubeconfig)
	if err != nil {
		return nil, "", echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Velero 클라이언트 생성
	service, err := clients.NewVeleroClientFromRawConfig(decodeKubeconfig)
	if err != nil {
		return nil, "", echo.NewHTTPError(http.StatusBadRequest, "Invalid kubeconfig: "+err.Error())
	}

	return service, namespace, nil
}

func (c *VeleroBodyController) determineNamespace(req *models.VeleroRequest, ctx echo.Context) string {
	// Priority: request body > query param > default
	if req.Namespace != "" {
		return req.Namespace
	}
	if ns := ctx.QueryParam("namespace"); ns != "" {
		return ns
	}
	return "velero" // default
}

// 공통 요청 처리 함수
func (c *VeleroBodyController) handleRequest(
	ctx echo.Context,
	operation string,
	handler func(context.Context, interfaces.VeleroService, string) (interface{}, error),
) error {
	start := time.Now()
	requestID := ctx.Response().Header().Get(echo.HeaderXRequestID)

	service, namespace, err := c.getVeleroServiceFromRequest(ctx)
	if err != nil {
		log.Printf("[ERROR] [%s] Failed to create service: %v", requestID, err)
		return err
	}

	result, err := handler(ctx.Request().Context(), service, namespace)
	if err != nil {
		duration := time.Since(start)
		log.Printf("[ERROR] [%s] Operation '%s' failed in %v: %v", requestID, operation, duration, err)

		return ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     err.Error(),
			Operation: operation,
			Namespace: namespace,
			Timestamp: time.Now(),
			RequestID: requestID,
			Duration:  duration.String(),
		})
	}

	duration := time.Since(start)
	log.Printf("[INFO] [%s] Operation '%s' completed in %v", requestID, operation, duration)

	return ctx.JSON(http.StatusOK, models.SuccessResponse{
		Data:      result,
		Success:   true,
		Operation: operation,
		Namespace: namespace,
		Timestamp: time.Now(),
		RequestID: requestID,
		Duration:  duration.String(),
	})

}

func (c *VeleroBodyController) GetBackups(ctx echo.Context) error {
	return c.handleRequest(ctx, "get_backups", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		backups, err := service.GetBackups(reqCtx, namespace)
		if err != nil {
			return nil, err
		}

		// Sort backups by creation time (newest first)
		sort.Slice(backups, func(i, j int) bool {
			return backups[i].CreationTimestamp.Time.After(backups[j].CreationTimestamp.Time)
		})

		// Add summary information
		summary := c.generateBackupSummary(backups)

		return map[string]interface{}{
			"backups": backups,
			"summary": summary,
		}, nil
	})
}

func (c *VeleroBodyController) GetRestores(ctx echo.Context) error {
	return c.handleRequest(ctx, "get_restores", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		restores, err := service.GetRestores(reqCtx, namespace)
		if err != nil {
			return nil, err
		}

		// Sort restores by creation time (newest first)
		sort.Slice(restores, func(i, j int) bool {
			return restores[i].CreationTimestamp.Time.After(restores[j].CreationTimestamp.Time)
		})

		// Add summary information
		summary := c.generateRestoreSummary(restores)

		return map[string]interface{}{
			"restores": restores,
			"summary":  summary,
		}, nil
	})
}

func (c *VeleroBodyController) GetBackupRepositories(ctx echo.Context) error {
	return c.handleRequest(ctx, "get_backup_repositories", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetBackupRepositories(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetBackupStorageLocations(ctx echo.Context) error {
	return c.handleRequest(ctx, "get_backup_storage_locations", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		locations, err := service.GetBackupStorageLocations(reqCtx, namespace)
		if err != nil {
			return nil, err
		}

		// Add status summary
		summary := c.generateStorageLocationSummary(locations)

		return map[string]interface{}{
			"locations": locations,
			"summary":   summary,
		}, nil
	})
}

func (c *VeleroBodyController) GetVolumeSnapshotLocations(ctx echo.Context) error {
	return c.handleRequest(ctx, "get_volume_snapshot_locations", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetVolumeSnapshotLocations(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetPodVolumeRestores(ctx echo.Context) error {
	return c.handleRequest(ctx, "get_pod_volume_restores", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetPodVolumeRestores(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetDownloadRequests(ctx echo.Context) error {
	return c.handleRequest(ctx, "get_download_requests", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetDownloadRequests(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetDataUploads(ctx echo.Context) error {
	return c.handleRequest(ctx, "get_data_uploads", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetDataUploads(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetDataDownloads(ctx echo.Context) error {
	return c.handleRequest(ctx, "get_data_downloads", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetDataDownloads(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetServerStatusRequests(ctx echo.Context) error {
	return c.handleRequest(ctx, "get_server_status_requests", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetServerStatusRequests(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetBackupDetails(ctx echo.Context) error {
	backupName := ctx.Param("name")
	if backupName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Backup name is required")
	}

	return c.handleRequest(ctx, "get_backup_details", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		backups, err := service.GetBackups(reqCtx, namespace)
		if err != nil {
			return nil, err
		}

		// Find specific backup
		var targetBackup *velerov1.Backup
		for i, backup := range backups {
			if backup.Name == backupName {
				targetBackup = &backups[i]
				break
			}
		}

		if targetBackup == nil {
			return nil, fmt.Errorf("backup '%s' not found", backupName)
		}

		// Generate detailed information
		details := c.generateBackupDetails(targetBackup)

		return details, nil
	})
}

func (c *VeleroBodyController) ValidateBackup(ctx echo.Context) error {
	var req models.BackupValidationRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	return c.handleRequest(ctx, "validate_backup", func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		backups, err := service.GetBackups(reqCtx, namespace)
		if err != nil {
			return nil, err
		}

		// Find specific backup
		var targetBackup *velerov1.Backup
		for i, backup := range backups {
			if backup.Name == req.BackupName {
				targetBackup = &backups[i]
				break
			}
		}

		if targetBackup == nil {
			return nil, fmt.Errorf("backup '%s' not found", req.BackupName)
		}

		// Comprehensive validation
		validation := c.validateBackupComprehensive(targetBackup)

		return models.BackupValidationResponse{
			BackupName:      req.BackupName,
			IsValid:         validation.IsValid,
			Warnings:        validation.Warnings,
			Errors:          validation.Errors,
			Status:          string(targetBackup.Status.Phase),
			CreatedAt:       targetBackup.CreationTimestamp.Time,
			CompletedAt:     targetBackup.Status.CompletionTimestamp,
			Size:            c.formatBytes(targetBackup.Status.Progress.TotalBytes),
			ExpiresAt:       targetBackup.Status.Expiration,
			Timestamp:       time.Now(),
			Recommendations: validation.Recommendations,
		}, nil
	})
}

func (c *VeleroBodyController) generateBackupDetails(backup *velerov1.Backup) models.BackupDetails {
	details := models.BackupDetails{
		Name:        backup.Name,
		Namespace:   backup.Namespace,
		Status:      string(backup.Status.Phase),
		CreatedAt:   backup.CreationTimestamp.Time,
		CompletedAt: backup.Status.CompletionTimestamp,
		ExpiresAt:   backup.Status.Expiration,
		Size:        c.formatBytes(backup.Status.Progress.TotalBytes),
		Progress:    backup.Status.Progress,
	}

	// Include/Exclude resources
	if backup.Spec.IncludedNamespaces != nil {
		details.IncludedNamespaces = backup.Spec.IncludedNamespaces
	}
	if backup.Spec.ExcludedNamespaces != nil {
		details.ExcludedNamespaces = backup.Spec.ExcludedNamespaces
	}
	if backup.Spec.IncludedResources != nil {
		details.IncludedResources = backup.Spec.IncludedResources
	}
	if backup.Spec.ExcludedResources != nil {
		details.ExcludedResources = backup.Spec.ExcludedResources
	}

	// Storage location
	details.StorageLocation = backup.Spec.StorageLocation

	// Labels and annotations
	details.Labels = backup.Labels
	details.Annotations = backup.Annotations

	// Validation errors
	details.ValidationErrors = backup.Status.ValidationErrors

	return details
}

func (c *VeleroBodyController) validateBackupComprehensive(backup *velerov1.Backup) *models.ValidationResult {
	result := &models.ValidationResult{IsValid: true}

	// Check backup phase
	switch backup.Status.Phase {
	case velerov1.BackupPhaseCompleted:
		// Good
	case velerov1.BackupPhaseFailed:
		result.IsValid = false
		result.Errors = append(result.Errors, "Backup failed")
	case velerov1.BackupPhasePartiallyFailed:
		result.Warnings = append(result.Warnings, "Backup partially failed - some resources may not be backed up")
		result.Recommendations = append(result.Recommendations, "Review backup logs to identify failed resources")
	case velerov1.BackupPhaseInProgress:
		result.Warnings = append(result.Warnings, "Backup still in progress")
	}

	// Check for validation errors
	if len(backup.Status.ValidationErrors) > 0 {
		result.IsValid = false
		for _, err := range backup.Status.ValidationErrors {
			result.Errors = append(result.Errors, err)
		}
		result.Recommendations = append(result.Recommendations, "Fix validation errors before using this backup")
	}

	// Check expiration
	if backup.Status.Expiration != nil {
		if backup.Status.Expiration.Before(time.Now()) {
			result.Warnings = append(result.Warnings, "Backup has expired and may be deleted")
			result.Recommendations = append(result.Recommendations, "Create a new backup if needed")
		} else if time.Until(*backup.Status.Expiration) < 24*time.Hour {
			result.Warnings = append(result.Warnings, "Backup will expire within 24 hours")
			result.Recommendations = append(result.Recommendations, "Consider extending backup retention if needed")
		}
	}

	// Check backup age
	age := time.Since(backup.CreationTimestamp.Time)
	if age > 30*24*time.Hour { // 30 days
		result.Warnings = append(result.Warnings, "Backup is older than 30 days")
		result.Recommendations = append(result.Recommendations, "Verify backup integrity before using for restoration")
	}

	// Check backup size
	if backup.Status.Progress.TotalBytes == 0 {
		result.Warnings = append(result.Warnings, "Backup size is 0 bytes - may indicate no data was backed up")
		result.Recommendations = append(result.Recommendations, "Verify backup configuration and included resources")
	}

	return result
}

// Helper methods for summary generation
func (c *VeleroBodyController) generateBackupSummary(backups []velerov1.Backup) models.BackupSummary {
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
		if backup.Status.Expiration != nil && backup.Status.Expiration.Before(time.Now()) {
			summary.Expired++
		}
	}

	return summary
}

func (c *VeleroBodyController) generateRestoreSummary(restores []velerov1.Restore) models.RestoreSummary {
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

func (c *VeleroBodyController) generateStorageLocationSummary(locations []velerov1.BackupStorageLocation) models.StorageLocationSummary {
	summary := models.StorageLocationSummary{
		Total: len(locations),
	}

	for _, location := range locations {
		switch location.Status.Phase {
		case velerov1.BackupStorageLocationPhaseAvailable:
			summary.Available++
		case velerov1.BackupStorageLocationPhaseUnavailable:
			summary.Unavailable++
		}
	}

	return summary
}

func (c *VeleroBodyController) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// 대상 클러스터 검증
func (c *VeleroBodyController) ValidateDestination(ctx echo.Context) error {
	var req models.VeleroRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if req.DestinationKubeconfig == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "destinationKubeconfig is required")
	}

	// 대상 클러스터 서비스 생성
	destService, err := clients.NewVeleroClientFromRawConfig(req.DestinationKubeconfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid destination kubeconfig: "+err.Error())
	}

	namespace := req.Namespace
	if namespace == "" {
		namespace = "velero"
	}

	// Velero 설치 상태 확인
	_, err = destService.GetBackupStorageLocations(ctx.Request().Context(), namespace)
	if err != nil {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"valid":   false,
			"error":   err.Error(),
			"message": "Destination cluster validation failed",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"valid":   true,
		"message": "Destination cluster is ready for Velero operations",
	})
}

// 스토리지 클래스 비교
func (c *VeleroBodyController) CompareStorageClasses(ctx echo.Context) error {
	var req models.VeleroRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if req.SourceKubeconfig == "" || req.DestinationKubeconfig == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Both source and destination kubeconfigs are required")
	}

	// 소스와 대상 클러스터의 스토리지 클래스를 가져오는 로직
	// 여기서는 간단한 예시로 구현
	sourceService, err := clients.NewVeleroClientFromRawConfig(req.SourceKubeconfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid source kubeconfig: "+err.Error())
	}

	destService, err := clients.NewVeleroClientFromRawConfig(req.DestinationKubeconfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid destination kubeconfig: "+err.Error())
	}

	namespace := req.Namespace
	if namespace == "" {
		namespace = "velero"
	}

	// 간단한 비교 로직 (실제로는 더 복잡한 구현 필요)
	sourceLocations, err := sourceService.GetBackupStorageLocations(ctx.Request().Context(), namespace)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get source storage locations: "+err.Error())
	}

	destLocations, err := destService.GetBackupStorageLocations(ctx.Request().Context(), namespace)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get destination storage locations: "+err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"sourceStorageLocations":      len(sourceLocations),
		"destinationStorageLocations": len(destLocations),
		"compatible":                  len(sourceLocations) > 0 && len(destLocations) > 0,
		"message":                     "Storage location comparison completed",
	})
}
