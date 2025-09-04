package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"taking.kr/velero/pkg/cache"
	"taking.kr/velero/pkg/handlers"
	"taking.kr/velero/pkg/health"
	"taking.kr/velero/pkg/utils"
)

// RegisterRoutes : 라우트 등록
func RegisterRoutes(e *echo.Echo, appCache *cache.Cache, workerPool *utils.WorkerPool, healthManager *health.HealthManager) {
	minioHandler := handlers.NewMinioHandler(appCache, workerPool, healthManager)
	helmHandler := handlers.NewHelmHandler(appCache, workerPool, healthManager)
	kubeHandler := handlers.NewKubernetesHandler(appCache, workerPool, healthManager)
	veleroHandler := handlers.NewVeleroHandler(appCache, workerPool, healthManager)

	api := e.Group("/api/v1")

	minio := api.Group("/minio")
	{
		minio.GET("/health", minioHandler.HealthCheck)
		minio.POST("/bucket_check", minioHandler.CreateBucketIfNotExists)
	}

	helm := api.Group("/helm")
	{
		helm.GET("/health", helmHandler.HealthCheck)
		helm.POST("/chart_check", helmHandler.IsChartInstalled)
	}

	kube := api.Group("/kube")
	{
		kube.GET("/health", kubeHandler.HealthCheck)
		kube.GET("/pods", kubeHandler.GetPods)
		kube.GET("/storage-classes", kubeHandler.GetStorageClasses)
	}

	velero := api.Group("/velero")
	{
		velero.GET("/health", veleroHandler.HealthCheck)
		velero.GET("/backups", veleroHandler.GetBackups)
		velero.GET("/restores", veleroHandler.GetRestores)
		velero.GET("/backup-repositories", veleroHandler.GetBackupRepositories)
		velero.GET("/backup-storage-locations", veleroHandler.GetBackupStorageLocations)
		velero.GET("/volume-snapshot-locations", veleroHandler.GetVolumeSnapshotLocations)
		//velero.GET("/pod-volume-restores", veleroHandler.GetPodVolumeRestores)
		//velero.GET("/download-requests", veleroHandler.GetDownloadRequests)
		//velero.GET("/data-uploads", veleroHandler.GetDataUploads)
		//velero.GET("/data-downloads", veleroHandler.GetDataDownloads)
		//velero.GET("/server-status-requests", veleroHandler.GetServerStatusRequests)
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

	// 전체 시스템 헬스체크
	api.GET("/health", func(c echo.Context) error {
		overallHealth := healthManager.OverallHealth(c.Request().Context())

		// 기존 응답 형식 유지하면서 새로운 정보 추가
		response := map[string]interface{}{
			"status":  overallHealth.Status,
			"version": "v1",
			"message": overallHealth.Message,
			"features": []string{
				"single-cluster",
				"multi-cluster",
				"backup-migration",
				"storage-validation",
				"caching",
				"worker-pool",
			},
			"timestamp": overallHealth.Timestamp,
			"services":  healthManager.GetAllCached(),
		}

		statusCode := http.StatusOK
		if overallHealth.Status != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}

		return c.JSON(statusCode, response)
	})

	// 개별 서비스 헬스체크
	api.GET("/health/:service", func(c echo.Context) error {
		serviceName := c.Param("service")
		serviceHealth := healthManager.CheckSingle(c.Request().Context(), serviceName)

		statusCode := http.StatusOK
		if serviceHealth.Status != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}

		return c.JSON(statusCode, serviceHealth)
	})
}
