package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	openapidocs "github.com/kohkimakimoto/echo-openapidocs"
	"github.com/labstack/echo/v4"

	"github.com/taking/kubemigrate/docs/swagger"
	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/logger"
	"github.com/taking/kubemigrate/internal/server"
)

// @title KubeMigrate API Server
// @version 1.0
// @description Kubernetes cluster migration and backup validation API with multi-cluster support

// @contact.name API Support
// @contact.url github.com/taking/kubemigrate/issues
// @contact.email taking@duck.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Config 불러오기
	cfg := config.Load()

	// Config 기반으로 Logger 초기화
	if err := logger.Init(logger.Config{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		OutputPath: "stdout",
	}); err != nil {
		panic("failed to logger initialize: " + err.Error())
	}

	logger.Info("KubeMigrate API Server starting")

	// 새로운 라우터 생성 (미들웨어 포함)
	e := server.NewRouter()
	e.HideBanner = true
	e.HidePort = true

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
	startServer(server, cfg)
}

// setupScalarDocs : ScalarDocs 설정
func setupScalarDocs(e *echo.Echo, cfg *config.Config) {
	// SwaggerInfo의 Host를 동적으로 설정
	swagger.SwaggerInfo.Host = cfg.Server.Host + ":" + cfg.Server.Port

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
		logger.String("url", "http://"+cfg.Server.Host+":"+cfg.Server.Port+"/docs"))
}

// startServer : 서버 실행
func startServer(server *http.Server, cfg *config.Config) {
	go func() {
		logger.Info("KubeMigrate API Server starting",
			logger.String("addr", cfg.Server.Host+":"+cfg.Server.Port),
			logger.String("docs_url", "http://"+cfg.Server.Host+":"+cfg.Server.Port+"/docs"),
			logger.String("health_url", "http://"+cfg.Server.Host+":"+cfg.Server.Port+"/api/v1/health"))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Server failed to start", logger.ErrorAttr(err))
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
		logger.Error("Server forced to shutdown", logger.ErrorAttr(err))
	}

	logger.Info("Server exited gracefully")
}
