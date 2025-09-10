package server

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/api/helm"
	"github.com/taking/kubemigrate/internal/api/kubernetes"
	"github.com/taking/kubemigrate/internal/api/minio"
	"github.com/taking/kubemigrate/internal/api/velero"
	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/handler"
	appMiddleware "github.com/taking/kubemigrate/internal/middleware"
	"github.com/taking/kubemigrate/pkg/utils"
)

// NewRouter 새로운 라우터를 생성합니다
func NewRouter(cfg *config.Config) *echo.Echo {
	e := echo.New()

	// 고급 미들웨어 설정 적용
	appMiddleware.SetupMiddleware(e, cfg)

	// 공통 컴포넌트 초기화
	workerPool := utils.NewWorkerPool(10) // 10개 워커

	// BaseHandler 생성
	baseHandler := handler.NewBaseHandler(workerPool)

	// 백그라운드 캐시 정리 작업 시작 (1분마다)
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			baseHandler.CleanupCache()
		}
	}()

	// 기능별 핸들러 생성
	veleroHandler := velero.NewHandler(baseHandler)
	helmHandler := helm.NewHandler(baseHandler)
	kubernetesHandler := kubernetes.NewHandler(baseHandler)
	minioHandler := minio.NewHandler(baseHandler)

	// 루트 라우트 테스트
	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"message": "KubeMigrate API Server - Go Standard Structure",
			"version": "1.0",
			"status":  "running",
		})
	})

	// API 그룹 설정
	api := e.Group("/api/v1")

	// Velero 관련 라우트 (백업 + 복구 통합)
	veleroGroup := api.Group("/velero")
	veleroGroup.POST("/health", veleroHandler.HealthCheck)
	// 백업 관련
	veleroGroup.POST("/backups", veleroHandler.GetBackups)
	veleroGroup.GET("/repositories", veleroHandler.GetBackupRepositories)
	veleroGroup.GET("/storage-locations", veleroHandler.GetBackupStorageLocations)
	// 복구 관련
	veleroGroup.POST("/restores", veleroHandler.GetRestores)
	veleroGroup.GET("/volume-snapshot-locations", veleroHandler.GetVolumeSnapshotLocations)
	veleroGroup.GET("/pod-volume-restores", veleroHandler.GetPodVolumeRestores)

	// Helm 관련 라우트 (RESTful)
	helmGroup := api.Group("/helm")
	helmGroup.POST("/health", helmHandler.HealthCheck)                  // 1. Helm 연결 테스트
	helmGroup.POST("/charts", helmHandler.InstallChart)                 // 2. 차트 설치
	helmGroup.GET("/charts", helmHandler.GetCharts)                     // 3. 차트 목록 조회
	helmGroup.GET("/charts/:name", helmHandler.GetChart)                // 4. 차트 상세 조회
	helmGroup.GET("/charts/:name/status", helmHandler.IsChartInstalled) // 6. 차트 설치 상태
	helmGroup.PUT("/charts/:name", helmHandler.UpgradeChart)            // 5. 차트 업그레이드
	helmGroup.GET("/charts/:name/history", helmHandler.GetChartHistory) // 8. 차트 히스토리 조회
	helmGroup.GET("/charts/:name/values", helmHandler.GetChartValues)   // 9. 차트 값 조회
	helmGroup.DELETE("/charts/:name", helmHandler.UninstallChart)       // 10. 차트 제거

	// Kubernetes 관련 라우트
	k8sGroup := api.Group("/kubernetes")
	k8sGroup.POST("/health", kubernetesHandler.HealthCheck)

	// 통합 API (조회)
	k8sGroup.GET("/:kind", kubernetesHandler.GetResources)       // 리스트 조회
	k8sGroup.GET("/:kind/:name", kubernetesHandler.GetResources) // 단일 조회
	// MinIO 관련 라우트
	minioGroup := api.Group("/minio")
	minioGroup.POST("/health", minioHandler.HealthCheck) // 1. 연결 상태 조회

	// 버킷 관리 (RESTful)
	minioGroup.GET("/buckets/:bucket", minioHandler.CheckBucketExists) // 2. 버킷 존재 확인
	minioGroup.POST("/buckets/:bucket", minioHandler.CreateBucket)     // 3. 버킷 생성
	minioGroup.GET("/buckets", minioHandler.ListBuckets)               // 4. 버킷 목록 조회
	minioGroup.DELETE("/buckets/:bucket", minioHandler.DeleteBucket)   // 5. 버킷 삭제

	// 객체 관리 (RESTful)
	minioGroup.GET("/buckets/:bucket/objects", minioHandler.ListObjects)                                          // 6. 객체 목록 조회
	minioGroup.POST("/buckets/:bucket/objects/*", minioHandler.PutObject)                                         // 7. 객체 업로드
	minioGroup.GET("/buckets/:bucket/objects/*", minioHandler.GetObject)                                          // 8. 객체 다운로드
	minioGroup.GET("/buckets/:bucket/objects/*", minioHandler.StatObject)                                         // 9.객체 정보 조회
	minioGroup.POST("/buckets/:srcBucket/objects/:srcObject/copy/:dstBucket/:dstObject", minioHandler.CopyObject) // 10. 객체 복사
	minioGroup.DELETE("/buckets/:bucket/objects/*", minioHandler.DeleteObject)                                    // 11. 객체 삭제

	// Presigned URL (RESTful) - 더 구체적인 라우트로 변경
	minioGroup.GET("/buckets/:bucket/objects/:object/presigned-get", minioHandler.PresignedGetObject) // 12. Presigned GET URL 생성
	minioGroup.PUT("/buckets/:bucket/objects/:object/presigned-put", minioHandler.PresignedPutObject) // 13. Presigned PUT URL 생성

	// 헬스체크 라우트 - API 서버 상태만 확인
	healthGroup := api.Group("/health")
	healthGroup.GET("", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status":    "healthy",
			"message":   "API server is running",
			"timestamp": time.Now(),
		})
	})

	// 캐시 관리 라우트
	cacheGroup := api.Group("/cache")
	cacheGroup.GET("/stats", func(c echo.Context) error {
		stats := baseHandler.GetCacheStats()
		return c.JSON(200, map[string]interface{}{
			"status": "success",
			"data":   stats,
		})
	})
	cacheGroup.POST("/cleanup", func(c echo.Context) error {
		baseHandler.CleanupCache()
		return c.JSON(200, map[string]interface{}{
			"status":  "success",
			"message": "Cache cleanup completed",
		})
	})

	return e
}
