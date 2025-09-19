// Package routes 헬스체크 및 시스템 상태 관련 라우트를 관리합니다.
package routes

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
)

// SetupHealthRoutes 헬스체크 및 시스템 상태 라우트를 설정합니다.
func SetupHealthRoutes(e *echo.Echo, baseHandler *handler.BaseHandler) {
	api := e.Group("/api/v1")

	// 헬스체크 라우트
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
	cacheGroup.GET("/detailed", func(c echo.Context) error {
		detailedStats := baseHandler.GetDetailedCacheStats()
		return c.JSON(200, map[string]interface{}{
			"status": "success",
			"data":   detailedStats,
		})
	})
	cacheGroup.GET("/cleanup", func(c echo.Context) error {
		baseHandler.CleanupCache()
		return c.JSON(200, map[string]interface{}{
			"status":  "success",
			"message": "Cache cleanup completed",
		})
	})
	cacheGroup.DELETE("/clean/:cache_key", func(c echo.Context) error {
		cacheKey := c.Param("cache_key")
		if cacheKey == "" {
			return c.JSON(400, map[string]interface{}{
				"status":  "error",
				"message": "Cache key is required",
			})
		}

		removed := baseHandler.CleanCacheByKey(cacheKey)
		if removed {
			return c.JSON(200, map[string]interface{}{
				"status":  "success",
				"message": "Cache item removed successfully",
				"data": map[string]interface{}{
					"cache_key": cacheKey,
					"removed":   true,
				},
			})
		}

		return c.JSON(404, map[string]interface{}{
			"status":  "error",
			"message": "Cache item not found",
			"data": map[string]interface{}{
				"cache_key": cacheKey,
				"removed":   false,
			},
		})
	})

	// 메모리 관리 라우트 (구현 예정)
	memoryGroup := api.Group("/memory")
	memoryGroup.GET("/stats", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status": "success",
			"data":   map[string]interface{}{"message": "Memory stats not implemented yet"},
		})
	})
	memoryGroup.POST("/optimize", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status":  "success",
			"message": "Memory optimization not implemented yet",
		})
	})
	memoryGroup.GET("/usage", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"usage_percent": 0,
				"is_high":       false,
			},
		})
	})
}
