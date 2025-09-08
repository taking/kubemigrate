package server

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/taking/kubemigrate/internal/api/helm"
	"github.com/taking/kubemigrate/internal/api/kubernetes"
	"github.com/taking/kubemigrate/internal/api/minio"
	"github.com/taking/kubemigrate/internal/api/velero"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/utils"
)

// ServiceHealth 개별 서비스 헬스 상태
type ServiceHealth struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// HealthSummary 헬스 체크 요약 정보
type HealthSummary struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Unhealthy int `json:"unhealthy"`
}

// HealthCheckResponse 통합 헬스 체크 응답
type HealthCheckResponse struct {
	Status    string                   `json:"status"`
	Message   string                   `json:"message"`
	Services  map[string]ServiceHealth `json:"services"`
	Summary   HealthSummary            `json:"summary"`
	Timestamp time.Time                `json:"timestamp"`
}

// NewRouter 새로운 라우터를 생성합니다
func NewRouter() *echo.Echo {
	e := echo.New()

	// 기본 미들웨어 설정
	e.Pre(middleware.RemoveTrailingSlash()) // 모든 요청에서 URL 뒤에 붙은 / 제거
	e.Use(middleware.Logger())              // 요청/응답 로그 출력
	e.Use(middleware.Recover())             // panic 발생 시 서버가 죽지 않고 500 에러로 복구
	e.Use(middleware.CORS())                // 기본 CORS 허용 설정

	// 공통 컴포넌트 초기화
	workerPool := utils.NewWorkerPool(10) // 10개 워커

	// BaseHandler 생성
	baseHandler := handler.NewBaseHandler(workerPool)

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

	// Helm 관련 라우트
	helmGroup := api.Group("/helm")
	helmGroup.POST("/health", helmHandler.HealthCheck)
	helmGroup.GET("/charts", helmHandler.GetCharts)
	helmGroup.GET("/chart/:name", helmHandler.GetChart)
	helmGroup.GET("/chart/:name/status", helmHandler.IsChartInstalled)
	helmGroup.DELETE("/chart/:name", helmHandler.UninstallChart)

	// Kubernetes 관련 라우트
	k8sGroup := api.Group("/kubernetes")
	k8sGroup.POST("/health", kubernetesHandler.HealthCheck)

	// 통합 API (조회)
	k8sGroup.GET("/:kind", kubernetesHandler.GetResources)       // 리스트 조회
	k8sGroup.GET("/:kind/:name", kubernetesHandler.GetResources) // 단일 조회

	// MinIO 관련 라우트
	minioGroup := api.Group("/minio")
	minioGroup.POST("/health", minioHandler.HealthCheck)
	minioGroup.POST("/bucket/exists", minioHandler.CheckBucketExists)
	minioGroup.POST("/bucket/create", minioHandler.CreateBucket)
	minioGroup.POST("/bucket/create-if-not-exists", minioHandler.CreateBucketIfNotExists)

	// 헬스체크 라우트
	healthGroup := api.Group("/health")
	healthGroup.GET("", func(c echo.Context) error {
		// 각 서비스별 health check 결과를 수집
		services := map[string]ServiceHealth{
			"kubernetes": checkKubernetesHealth(c),
			"minio":      checkMinioHealth(c),
			"helm":       checkHelmHealth(c),
			"velero":     checkVeleroHealth(c),
		}

		// 전체 상태 계산
		healthyCount := 0
		totalCount := len(services)
		overallStatus := "healthy"

		for _, service := range services {
			if service.Status == "healthy" {
				healthyCount++
			} else {
				overallStatus = "unhealthy"
			}
		}

		statusCode := 200
		if overallStatus != "healthy" {
			statusCode = 503
		}

		// 깔끔한 응답 구조
		response := HealthCheckResponse{
			Status:   overallStatus,
			Message:  fmt.Sprintf("Services: %d/%d healthy", healthyCount, totalCount),
			Services: services,
			Summary: HealthSummary{
				Total:     totalCount,
				Healthy:   healthyCount,
				Unhealthy: totalCount - healthyCount,
			},
			Timestamp: time.Now(),
		}

		return c.JSON(statusCode, response)
	})

	return e
}

// checkKubernetesHealth Kubernetes 서비스 health check
func checkKubernetesHealth(c echo.Context) ServiceHealth {
	// 기본 클라이언트로 health check 수행
	k8sClient := client.NewClient().Kubernetes()
	if k8sClient == nil {
		return ServiceHealth{
			Status:    "unhealthy",
			Message:   "Kubernetes client not initialized",
			Timestamp: time.Now(),
			Error:     "client not available",
		}
	}

	// 간단한 연결 테스트
	_, err := k8sClient.GetPods(c.Request().Context(), "default", "")
	if err != nil {
		return ServiceHealth{
			Status:    "unhealthy",
			Message:   "Kubernetes connection failed",
			Timestamp: time.Now(),
			Error:     err.Error(),
		}
	}

	return ServiceHealth{
		Status:    "healthy",
		Message:   "Kubernetes connection is working",
		Timestamp: time.Now(),
	}
}

// checkMinioHealth MinIO 서비스 health check
func checkMinioHealth(c echo.Context) ServiceHealth {
	// 기본 클라이언트로 health check 수행
	minioClient := client.NewClient().Minio()
	if minioClient == nil {
		return ServiceHealth{
			Status:    "unhealthy",
			Message:   "MinIO client not initialized",
			Timestamp: time.Now(),
			Error:     "client not available",
		}
	}

	// 간단한 연결 테스트
	_, err := minioClient.ListBuckets(c.Request().Context())
	if err != nil {
		return ServiceHealth{
			Status:    "unhealthy",
			Message:   "MinIO connection failed",
			Timestamp: time.Now(),
			Error:     err.Error(),
		}
	}

	return ServiceHealth{
		Status:    "healthy",
		Message:   "MinIO connection is working",
		Timestamp: time.Now(),
	}
}

// checkHelmHealth Helm 서비스 health check
func checkHelmHealth(c echo.Context) ServiceHealth {
	// 기본 클라이언트로 health check 수행
	helmClient := client.NewClient().Helm()
	if helmClient == nil {
		return ServiceHealth{
			Status:    "unhealthy",
			Message:   "Helm client not initialized",
			Timestamp: time.Now(),
			Error:     "client not available",
		}
	}

	// 간단한 연결 테스트
	_, err := helmClient.GetCharts(c.Request().Context(), "default")
	if err != nil {
		return ServiceHealth{
			Status:    "unhealthy",
			Message:   "Helm connection failed",
			Timestamp: time.Now(),
			Error:     err.Error(),
		}
	}

	return ServiceHealth{
		Status:    "healthy",
		Message:   "Helm connection is working",
		Timestamp: time.Now(),
	}
}

// checkVeleroHealth Velero 서비스 health check
func checkVeleroHealth(c echo.Context) ServiceHealth {
	// 기본 클라이언트로 health check 수행
	veleroClient := client.NewClient().Velero()
	if veleroClient == nil {
		return ServiceHealth{
			Status:    "unhealthy",
			Message:   "Velero client not initialized",
			Timestamp: time.Now(),
			Error:     "client not available",
		}
	}

	// 간단한 연결 테스트
	_, err := veleroClient.GetBackups(c.Request().Context(), "velero")
	if err != nil {
		return ServiceHealth{
			Status:    "unhealthy",
			Message:   "Velero connection failed",
			Timestamp: time.Now(),
			Error:     err.Error(),
		}
	}

	return ServiceHealth{
		Status:    "healthy",
		Message:   "Velero connection is working",
		Timestamp: time.Now(),
	}
}
