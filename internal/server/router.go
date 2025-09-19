// Package server HTTP 서버와 라우팅을 관리합니다.
package server

import (
	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/api/helm"
	"github.com/taking/kubemigrate/internal/api/kubernetes"
	"github.com/taking/kubemigrate/internal/api/minio"
	"github.com/taking/kubemigrate/internal/api/velero"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/job"
	appMiddleware "github.com/taking/kubemigrate/internal/middleware"
	"github.com/taking/kubemigrate/internal/server/routes"
	"github.com/taking/kubemigrate/pkg/config"
	"github.com/taking/kubemigrate/pkg/constants"
)

// NewRouter 새로운 라우터를 생성합니다
func NewRouter(cfg *config.Config) *echo.Echo {
	e := echo.New()

	// 미들웨어 설정
	appMiddleware.SetupMiddleware(e, cfg)

	// 공통 컴포넌트 초기화
	workerPool := job.NewWorkerPool(constants.DefaultWorkerPoolSize)

	// BaseHandler 생성
	baseHandler := handler.NewBaseHandler(workerPool)

	// 백그라운드 캐시 정리 작업 시작  (1분마다)
	StartBackgroundTasks(baseHandler)

	// 핸들러 생성
	handlers := createHandlers(baseHandler)

	// 라우트 설정
	setupRoutes(e, handlers)

	return e
}

// createHandlers 모든 핸들러를 생성합니다.
func createHandlers(baseHandler *handler.BaseHandler) *Handlers {
	return &Handlers{
		Velero:     velero.NewHandler(baseHandler),
		Helm:       helm.NewHandler(baseHandler),
		Kubernetes: kubernetes.NewHandler(baseHandler),
		Minio:      minio.NewHandler(baseHandler),
		Base:       baseHandler,
	}
}

// setupRoutes 모든 라우트를 설정합니다.
func setupRoutes(e *echo.Echo, handlers *Handlers) {
	// 루트 라우트
	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"message": "KubeMigrate API Server - Go Standard Structure",
			"version": "1.0",
			"status":  "running",
		})
	})

	// 서비스별 라우트 설정
	routes.SetupVeleroRoutes(e, handlers.Velero)
	routes.SetupHelmRoutes(e, handlers.Helm)
	routes.SetupKubernetesRoutes(e, handlers.Kubernetes)
	routes.SetupMinioRoutes(e, handlers.Minio)
	routes.SetupHealthRoutes(e, handlers.Base)
}

// Handlers 모든 핸들러를 포함하는 구조체입니다.
type Handlers struct {
	Velero     *velero.Handler
	Helm       *helm.Handler
	Kubernetes *kubernetes.Handler
	Minio      *minio.Handler
	Base       *handler.BaseHandler
}
