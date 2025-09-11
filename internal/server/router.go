package server

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/api/plugin"
	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/handler"
	appMiddleware "github.com/taking/kubemigrate/internal/middleware"
	"github.com/taking/kubemigrate/internal/response"
	pluginpkg "github.com/taking/kubemigrate/pkg/plugin"
	"github.com/taking/kubemigrate/pkg/utils"
)

// NewRouter 새로운 라우터를 생성합니다 (플러그인 아키텍처)
func NewRouter(cfg *config.Config, pluginManager *pluginpkg.Manager) *echo.Echo {
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

	// API 그룹 생성
	apiGroup := e.Group("/api/v1")

	// 플러그인 라우트 등록
	if err := pluginManager.RegisterAllRoutes(apiGroup); err != nil {
		panic("Failed to register plugin routes: " + err.Error())
	}

	// 플러그인 관리 API 등록
	pluginHandler := plugin.NewHandler(baseHandler, pluginManager)
	pluginGroup := apiGroup.Group("/plugins")
	pluginGroup.GET("", pluginHandler.ListPlugins)
	pluginGroup.GET("/:name", pluginHandler.GetPlugin)
	pluginGroup.GET("/health", pluginHandler.HealthCheckAllPlugins)
	pluginGroup.GET("/:name/health", pluginHandler.HealthCheckPlugin)

	// 헬스체크 엔드포인트
	e.GET("/api/v1/health", func(c echo.Context) error {
		healthData := map[string]interface{}{
			"status": "UP",
			"components": map[string]interface{}{
				"kubemigrate": map[string]interface{}{
					"status": "UP",
				},
			},
		}
		return response.RespondWithData(c, 200, healthData)
	})

	return e
}
