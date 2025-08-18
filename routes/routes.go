package routes

import (
	"github.com/labstack/echo/v4"
	ctrl "taking.kr/velero/controller"
)

func RegisterRoutes(e *echo.Echo) {
	veleroCtl := ctrl.VeleroController{}
	kubeCtl := ctrl.KubeController{}

	api := e.Group("/api/v1")

	// VeleroController
	api.GET("/velero/backups", veleroCtl.GetBackups)
	api.GET("/velero/restores", veleroCtl.GetRestores)
	api.GET("/velero/backup-repositories", veleroCtl.GetBackupRepositories)
	api.GET("/velero/backup-storage-locations", veleroCtl.GetBackupStorageLocations)
	api.GET("/velero/volume-snapshot-locations", veleroCtl.GetVolumeSnapshotLocations)
	api.GET("/velero/pod-volume-restores", veleroCtl.GetPodVolumeRestores)
	api.GET("/velero/download-requests", veleroCtl.GetDownloadRequests)
	api.GET("/velero/data-uploads", veleroCtl.GetDataUploads)
	api.GET("/velero/data-downloads", veleroCtl.GetDataDownloads)
	api.GET("/velero/server-status-requests", veleroCtl.GetServerStatusRequests)

	// KubeController
	api.GET("/kube/resources/:group/:version/:resource/:name", kubeCtl.GetResources)
}
