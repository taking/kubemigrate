// Package routes Kubernetes 관련 라우트를 관리합니다.
package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/api/kubernetes"
)

// SetupKubernetesRoutes Kubernetes 관련 라우트를 설정합니다.
func SetupKubernetesRoutes(e *echo.Echo, kubernetesHandler *kubernetes.Handler) {
	api := e.Group("/api/v1")
	k8sGroup := api.Group("/kubernetes")

	// 헬스체크
	k8sGroup.POST("/health", kubernetesHandler.HealthCheck)

	// 통합 리소스 조회 API
	k8sGroup.GET("/:kind", kubernetesHandler.GetResources)       // 리스트 조회
	k8sGroup.GET("/:kind/:name", kubernetesHandler.GetResources) // 단일 조회
}
