package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	openapidocs "github.com/kohkimakimoto/echo-openapidocs"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	_ "github.com/taking/kubemigrate/docs/swagger" // Scalar Docs Swagger
	"github.com/taking/kubemigrate/pkg/cache"
	"github.com/taking/kubemigrate/pkg/config"
	"github.com/taking/kubemigrate/pkg/health"
	"github.com/taking/kubemigrate/pkg/logger"
	"github.com/taking/kubemigrate/pkg/middleware"
	"github.com/taking/kubemigrate/pkg/router"
	"github.com/taking/kubemigrate/pkg/utils"
)

// @title KubeMigrate API Server
// @version 1.0
// @description Kubernetes cluster migration and backup validation API with multi-cluster support

// @contact.name API Support
// @contact.url github.com/taking/kubemigrate/issues
// @contact.email taking@duck.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:9091
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Logger 초기화
	if err := logger.InitDefault(); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	// Config 불러오기
	cfg := config.Load()

	// 캐시 시스템 초기화 (5분 TTL)
	appCache := cache.NewCache(5 * time.Minute)
	logger.Info("Cache system initialized", zap.Duration("ttl", 5*time.Minute))

	// 워커 풀 초기화 (CPU 코어 수 * 2)
	workerPool := utils.NewWorkerPool(runtime.NumCPU() * 2)
	defer workerPool.Close()
	logger.Info("Worker pool initialized", zap.Int("workers", runtime.NumCPU()*2))

	// 헬스체크 매니저 초기화 (30초 타임아웃)
	healthManager := health.NewHealthManager(30 * time.Second)
	logger.Info("Health manager initialized", zap.Duration("timeout", 30*time.Second))

	// Echo 인스턴스 생성
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// 기본 미들웨어 설정
	middleware.SetupMiddleware(e, cfg)

	// API 라우트 등록 (의존성 주입)
	router.RegisterRoutes(e, appCache, workerPool, healthManager)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      e,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Scalar API 문서 등록
	setupScalarDocs(e, cfg)

	// 서버 실행
	startServer(server)
}

// setupScalarDocs : ScalarDocs 설정
func setupScalarDocs(e *echo.Echo, cfg *config.Config) {
	// OpenAPI JSON 엔드포인트
	e.GET("/swagger.json", func(c echo.Context) error {
		return c.File("docs/swagger/swagger.json")
	})

	// Scalar API 문서 등록
	openapidocs.ScalarDocuments(e, "docs", openapidocs.ScalarConfig{
		SpecUrl: "/swagger.json",
		Title:   "KubeMigrate API Documentation",
		Theme:   "blue",
	})

	logger.Info("API Documentation available",
		zap.String("url", "http://localhost:"+cfg.Server.Port+"/docs"))
}

// startServer : 서버 실행
func startServer(server *http.Server) {
	go func() {
		logger.Info("KubeMigrate API Server starting",
			zap.String("addr", server.Addr),
			zap.String("docs_url", "http://localhost:9091/docs"),
			zap.String("health_url", "http://localhost:9091/api/v1/health"))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// SIGTERM, SIGINT 처리
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}
