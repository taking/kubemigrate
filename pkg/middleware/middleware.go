package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"runtime"
	"taking.kr/velero/pkg/config"
	"time"
)

// SetupMiddleware : Echo 서버에 공통 미들웨어 설정
// 로깅, 복구, CORS, 압축, 타임아웃, 레이트 제한 등을 구성
func SetupMiddleware(e *echo.Echo, cfg *config.Config) {
	// Request ID 생성 미들웨어
	e.Use(middleware.RequestID())

	// Logger 미들웨어: 요청 로그를 JSON 형식으로 출력
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}","level":"info","message":"${method} ${uri}","status":${status},"latency":"${latency_human}","request_id":"${id}"}` + "\n",
	}))

	// Recover 미들웨어: Panic 발생 시 서버 종료 방지 및 500 응답 반환
	e.Use(middleware.Recover())

	// CORS 설정: 모든 도메인 허용, GET/POST/PUT/DELETE/OPTIONS 메서드 허용
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))

	// Gzip 압축 미들웨어: 응답 데이터 압축
	e.Use(middleware.Gzip())

	// GOMAXPROCS 설정: CPU 코어 수만큼 최대 프로세스 사용
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 요청 타임아웃 미들웨어: cfg.Timeouts.Request 기준
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: cfg.Timeouts.Request,
	}))

	// 레이트 리미팅 미들웨어 (옵션)
	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper, // 조건에 따라 스킵 가능
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      100,         // 초당 허용 요청 수
				Burst:     50,          // 최대 버스트 수
				ExpiresIn: time.Minute, // 카운터 만료 시간
			},
		),
		// 요청자 식별자 추출 (IP 기준)
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		// 레이트 초과 시 응답
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(429, map[string]string{"error": "rate limit exceeded"})
		},
		// 레이트 초과 시 DenyHandler 호출
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(429, map[string]string{"error": "rate limit exceeded"})
		},
	}))
}
