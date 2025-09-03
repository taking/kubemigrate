package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/kohkimakimoto/echo-openapidocs"
	"github.com/labstack/echo/v4"

	"taking.kr/velero/config"
	_ "taking.kr/velero/docs" // Scalar Docs Swagger
	"taking.kr/velero/middleware"
	"taking.kr/velero/routes"
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

	// Config 불러오기
	cfg := config.Load()

	// Echo 인스턴스 생성
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// 기본 미들웨어 설정
	middleware.SetupMiddleware(e, cfg)

	// API 라우트 등록
	routes.RegisterRoutes(e)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      e,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Swagger Docs 생성
	if err := generateSwagger(); err != nil {
		log.Printf("failed to generate swagger: %v", err)
	} else {
		log.Println("Swagger.yaml / swagger.json updated")
	}

	// Scalar API 문서 등록
	setupScalarDocs(e, cfg)

	// 서버 실행
	startServer(server)
}

// generateSwagger : 서버 실행 시 Swagger 문서 자동 갱신
func generateSwagger() error {
	log.Println("Generating OpenAPI docs...")

	// swag CLI를 subprocess로 실행
	cmd := exec.Command("swag", "init", "-g", "main.go", "-o", "docs", "--parseDependency", "--parseInternal")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// setupScalarDocs : ScalarDocs 설정
func setupScalarDocs(e *echo.Echo, cfg *config.Config) {
	// OpenAPI JSON 엔드포인트
	e.GET("/swagger.json", func(c echo.Context) error {
		return c.File("./docs/swagger.json")
	})

	// Scalar API 문서 등록
	openapidocs.ScalarDocuments(e, "/docs", openapidocs.ScalarConfig{
		SpecUrl: "/swagger.json",
		Title:   "Velero API Documentation",
		Theme:   "blue",
	})

	log.Printf("📖 API Documentation available at: http://localhost:%s/docs",
		cfg.Server.Port)
}

// startServer : 서버 실행
func startServer(server *http.Server) {
	go func() {
		log.Printf("Velero API Server starting on %s", server.Addr)
		log.Printf("API Documentation: http://localhost:9091/docs")
		log.Printf("Health Check: http://localhost:9091/api/v1/health")

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// SIGTERM, SIGINT 처리
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
