// Package routes Helm 관련 라우트를 관리합니다.
package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/api/helm"
)

// SetupHelmRoutes Helm 관련 라우트를 설정합니다.
func SetupHelmRoutes(e *echo.Echo, helmHandler *helm.Handler) {
	api := e.Group("/api/v1")
	helmGroup := api.Group("/helm")

	// 헬스체크
	helmGroup.POST("/health", helmHandler.HealthCheck)

	// 차트 관리 라우트 (RESTful)
	helmGroup.POST("/charts", helmHandler.InstallChart)                 // 차트 설치
	helmGroup.GET("/charts", helmHandler.GetCharts)                     // 차트 목록 조회
	helmGroup.GET("/charts/:name", helmHandler.GetChart)                // 차트 상세 조회
	helmGroup.GET("/charts/:name/status", helmHandler.IsChartInstalled) // 차트 설치 상태
	helmGroup.PUT("/charts/:name", helmHandler.UpgradeChart)            // 차트 업그레이드
	helmGroup.GET("/charts/:name/history", helmHandler.GetChartHistory) // 차트 히스토리 조회
	helmGroup.GET("/charts/:name/values", helmHandler.GetChartValues)   // 차트 값 조회
	helmGroup.DELETE("/charts/:name", helmHandler.UninstallChart)       // 차트 제거
}
