package middleware

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/errors"
	"github.com/taking/kubemigrate/internal/logger"
)

// ErrorHandlerMiddleware 에러 처리 미들웨어
func ErrorHandlerMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	errorHandler := errors.NewErrorHandler(logger.GetLogger(), cfg.Logging.Level == "debug")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 요청 처리
			err := next(c)

			// 에러가 없는 경우
			if err == nil {
				return nil
			}

			// 에러 처리
			return errorHandler.HandleError(c, err, "API_REQUEST")
		}
	}
}

// RecoveryMiddleware 패닉 복구 미들웨어
func RecoveryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					// 패닉을 에러로 변환
					err := errors.NewInternalError("panic_recovery",
						errors.NewInternalError("panic", fmt.Errorf("%v", r)))

					// 에러 처리
					errorHandler := errors.NewErrorHandler(logger.GetLogger(), true)
					_ = errorHandler.HandleError(c, err, "PANIC_RECOVERY")
				}
			}()

			return next(c)
		}
	}
}
