// Package routes Velero 관련 라우트를 관리합니다.
package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/api/velero"
)

// SetupVeleroRoutes Velero 관련 라우트를 설정합니다.
func SetupVeleroRoutes(e *echo.Echo, veleroHandler *velero.Handler) {
	api := e.Group("/api/v1")
	veleroGroup := api.Group("/velero")

	// 헬스체크
	veleroGroup.POST("/health", veleroHandler.HealthCheck)

	// 설치 및 설정 라우트
	veleroGroup.POST("/install", veleroHandler.InstallVeleroWithMinIO)

	// 백업 관련 라우트
	veleroGroup.GET("/backups", veleroHandler.GetBackups)
	veleroGroup.GET("/backups/:backupName", veleroHandler.GetBackup)
	veleroGroup.POST("/backups", veleroHandler.CreateBackup)
	veleroGroup.DELETE("/backups/:backupName", veleroHandler.DeleteBackup)
	veleroGroup.POST("/backups/:backupName/validate", veleroHandler.ValidateBackup)

	// 복구 관련 라우트
	veleroGroup.GET("/restores", veleroHandler.GetRestores)
	veleroGroup.GET("/restores/:restoreName", veleroHandler.GetRestore)
	veleroGroup.POST("/restores", veleroHandler.CreateRestore)
	veleroGroup.DELETE("/restores/:restoreName", veleroHandler.DeleteRestore)
	veleroGroup.POST("/restores/:restoreName/validate", veleroHandler.ValidateRestore)

	// Velero 리소스 조회 라우트
	veleroGroup.GET("/repositories", veleroHandler.GetBackupRepositories)
	veleroGroup.GET("/storage-locations", veleroHandler.GetBackupStorageLocations)
	veleroGroup.GET("/volume-snapshot-locations", veleroHandler.GetVolumeSnapshotLocations)
	veleroGroup.GET("/pod-volume-restores", veleroHandler.GetPodVolumeRestores)

	// 비동기 작업 관리 라우트
	veleroGroup.GET("/status/:jobId", veleroHandler.GetJobStatus) // 작업 상태 조회
	veleroGroup.GET("/logs/:jobId", veleroHandler.GetJobLogs)     // 작업 로그 조회
	veleroGroup.GET("/jobs", veleroHandler.GetAllJobs)            // 모든 작업 조회
}
