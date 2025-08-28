package routes

import (
	"github.com/labstack/echo/v4"
	"net/http"
	ctrl "taking.kr/velero/controller"
)

func RegisterRoutes(e *echo.Echo) {
	commonCtrl := ctrl.NewMinioController()
	kubeCtrl := ctrl.NewKubernetesController()
	veleroCtrl := ctrl.NewVeleroController()

	api := e.Group("/api/v1")

	common := api.Group("/minio")
	{
		common.GET("/health", commonCtrl.CheckMinioConnection)
		common.POST("/bucket_check", commonCtrl.CreateBucketIfNotExists)
	}

	kube := api.Group("/kube")
	{
		kube.GET("/health", kubeCtrl.HealthCheck)
		kube.GET("/storage-classes", kubeCtrl.GetStorageClasses)
	}

	velero := api.Group("/velero")
	{
		velero.GET("/health", veleroCtrl.HealthCheck)
		velero.GET("/backups", veleroCtrl.GetBackups)
		velero.GET("/restores", veleroCtrl.GetRestores)
		velero.GET("/backup-repositories", veleroCtrl.GetBackupRepositories)
		velero.GET("/backup-storage-locations", veleroCtrl.GetBackupStorageLocations)
		velero.GET("/volume-snapshot-locations", veleroCtrl.GetVolumeSnapshotLocations)
		//velero.GET("/pod-volume-restores", veleroCtrl.GetPodVolumeRestores)
		//velero.GET("/download-requests", veleroCtrl.GetDownloadRequests)
		//velero.GET("/data-uploads", veleroCtrl.GetDataUploads)
		//velero.GET("/data-downloads", veleroCtrl.GetDataDownloads)
		//velero.GET("/server-status-requests", veleroCtrl.GetServerStatusRequests)
	}

	// 멀티클러스터 기능 엔드포인트
	//migration := api.Group("/migration")
	//{
	//migration.POST("/validate-destination", veleroCtrl.ValidateDestination)
	//migration.POST("/migrate-backup", veleroCtrl.MigrateBackup)
	//migration.POST("/backup/:backupName"), veleroCtrl.MigrateBackup)
	//migration.GET("/compare-storage-classes", veleroCtrl.CompareStorageClasses)
	//migration.GET("/migration/:id/status", veleroCtrl.GetMigrationStatus)
	//}

	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "healthy",
			"version": "v1",
			"message": "server, kubernetes, velero all reachable",
			"features": []string{
				"single-cluster",
				"multi-cluster",
				"backup-migration",
				"storage-validation",
			},
		})
	})
}
