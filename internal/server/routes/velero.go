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
	veleroGroup.POST("/backups", veleroHandler.GetBackups)
	veleroGroup.GET("/repositories", veleroHandler.GetBackupRepositories)
	veleroGroup.GET("/storage-locations", veleroHandler.GetBackupStorageLocations)

	// 복구 관련 라우트
	veleroGroup.POST("/restores", veleroHandler.GetRestores)
	veleroGroup.GET("/volume-snapshot-locations", veleroHandler.GetVolumeSnapshotLocations)
	veleroGroup.GET("/pod-volume-restores", veleroHandler.GetPodVolumeRestores)
}
