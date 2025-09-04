package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/kohkimakimoto/echo-openapidocs"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	_ "github.com/taking/velero/api/swagger" // Scalar Docs Swagger
	"github.com/taking/velero/internal/config"
	"github.com/taking/velero/internal/handler"
	"github.com/taking/velero/internal/middleware"
	"github.com/taking/velero/pkg"
	"github.com/taking/velero/pkg/logger"
)

// @title Velero API Server
// @version 1.0
// @description Velero backup and restore management API with multi-cluster support
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

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

	// Echo 인스턴스 생성
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// 기본 미들웨어 설정
	middleware.SetupMiddleware(e, cfg)

	// API 라우트 등록
	handler.RegisterRoutes(e)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      e,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Swagger Docs 생성
	if err := generateSwagger(); err != nil {
		logger.Error("failed to generate swagger", zap.Error(err))
	} else {
		logger.Info("Swagger.yaml / swagger.json updated")
	}

	// Scalar API 문서 등록
	setupScalarDocs(e, cfg)

	// 서버 실행
	startServer(server)
}

// generateSwagger : 서버 실행 시 Swagger 문서 자동 갱신
func generateSwagger() error {
	logger.Info("Generating OpenAPI docs...")

	// swag CLI를 subprocess로 실행
	cmd := exec.Command(
		"swag", "init",
		"-g", "main.go",
		"-o", "../../docs/swagger",
		"--parseDependency",
		"--parseInternal",
	)
	cmd.Dir = "cmd"

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// swagger.json / yaml 복사해서 docs/swagger/ 로 배포
	os.MkdirAll("docs/swagger", 0755)
	_ = pkg.CopyFile("api/swagger/swagger.json", "docs/swagger/swagger.json")
	_ = pkg.CopyFile("api/swagger/swagger.yaml", "docs/swagger/swagger.yaml")

	return nil
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
		Title:   "Velero API Documentation",
		Theme:   "blue",
	})

	logger.Info("API Documentation available",
		zap.String("url", "http://localhost:"+cfg.Server.Port+"/docs"))
}

// startServer : 서버 실행
func startServer(server *http.Server) {
	go func() {
		logger.Info("Velero API Server starting",
			zap.String("addr", server.Addr),
			zap.String("docs_url", "http://localhost:9091/api/swagger"),
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
