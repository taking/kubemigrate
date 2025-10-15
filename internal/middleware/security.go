package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// SecurityConfig : 보안 미들웨어 설정
type SecurityConfig struct {
	EnableCORS            bool
	EnableSecurityHeaders bool
	AllowedOrigins        []string
	MaxRequestSize        int64
}

// DefaultSecurityConfig : 기본 보안 설정
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EnableCORS:            true,
		EnableSecurityHeaders: true,
		AllowedOrigins:        []string{"*"},
		MaxRequestSize:        10 * 1024 * 1024, // 10MB
	}
}

// SecurityMiddleware : 보안 미들웨어
func SecurityMiddleware(config *SecurityConfig) echo.MiddlewareFunc {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 요청 크기 제한
			if config.MaxRequestSize > 0 {
				c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, config.MaxRequestSize)
			}

			// 보안 헤더 설정
			if config.EnableSecurityHeaders {
				setSecurityHeaders(c)
			}

			// CORS 설정
			if config.EnableCORS {
				setCORSHeaders(c, config.AllowedOrigins)
			}

			// 요청 검증
			if err := validateRequest(c); err != nil {
				return err
			}

			return next(c)
		}
	}
}

// setSecurityHeaders : 보안 헤더 설정
func setSecurityHeaders(c echo.Context) {
	headers := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'",
	}

	for key, value := range headers {
		c.Response().Header().Set(key, value)
	}
}

// setCORSHeaders : CORS 헤더 설정
func setCORSHeaders(c echo.Context, allowedOrigins []string) {
	origin := c.Request().Header.Get("Origin")

	// Origin 검증
	if isAllowedOrigin(origin, allowedOrigins) {
		c.Response().Header().Set("Access-Control-Allow-Origin", origin)
	}

	c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	c.Response().Header().Set("Access-Control-Max-Age", "86400")
}

// isAllowedOrigin : Origin이 허용된 목록에 있는지 확인
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// validateRequest : 요청 검증
func validateRequest(c echo.Context) error {
	// HTTP 메서드 검증
	if !isValidHTTPMethod(c.Request().Method) {
		return echo.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed")
	}

	// Content-Type 검증 (POST, PUT 요청의 경우)
	if requiresContentType(c.Request().Method) {
		contentType := c.Request().Header.Get("Content-Type")
		if !isValidContentType(contentType) {
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, "Invalid content type")
		}
	}

	// 경로 검증
	if !isValidPath(c.Request().URL.Path) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid path")
	}

	return nil
}

// isValidHTTPMethod : 유효한 HTTP 메서드인지 확인
func isValidHTTPMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"}
	for _, valid := range validMethods {
		if method == valid {
			return true
		}
	}
	return false
}

// requiresContentType : Content-Type이 필요한 HTTP 메서드인지 확인
func requiresContentType(method string) bool {
	return method == "POST" || method == "PUT"
}

// isValidContentType : 유효한 Content-Type인지 확인
func isValidContentType(contentType string) bool {
	validTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
	}

	for _, valid := range validTypes {
		if strings.HasPrefix(contentType, valid) {
			return true
		}
	}
	return false
}

// isValidPath : 유효한 경로인지 확인
func isValidPath(path string) bool {
	// 경로 주입 공격 방지
	if strings.Contains(path, "..") {
		return false
	}

	// 너무 긴 경로 방지
	if len(path) > 2048 {
		return false
	}

	// 특수 문자 검증
	if strings.ContainsAny(path, "<>\"'&") {
		return false
	}

	return true
}

// RateLimitMiddleware : 요청 제한 미들웨어
func RateLimitMiddleware(requestsPerMinute int) echo.MiddlewareFunc {
	// 간단한 메모리 기반 rate limiter
	// 실제 운영 환경에서는 Redis 등을 사용하는 것이 좋습니다
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// IP 기반 rate limiting 구현
			_ = c.RealIP() // 현재는 사용하지 않지만 향후 구현 예정

			// TODO: 실제 rate limiting 로직 구현
			// 현재는 기본 구현만 제공

			return next(c)
		}
	}
}

// InputSanitizationMiddleware : 입력 데이터 정화 미들웨어
func InputSanitizationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 요청 파라미터 정화
			sanitizeQueryParams(c)

			// 헤더 정화
			sanitizeHeaders(c)

			return next(c)
		}
	}
}

// sanitizeQueryParams : 쿼리 파라미터 정화
func sanitizeQueryParams(c echo.Context) {
	query := c.Request().URL.Query()
	for key, values := range query {
		for i, value := range values {
			// HTML 태그 제거
			value = strings.ReplaceAll(value, "<", "&lt;")
			value = strings.ReplaceAll(value, ">", "&gt;")
			value = strings.ReplaceAll(value, "\"", "&quot;")
			value = strings.ReplaceAll(value, "'", "&#x27;")
			value = strings.ReplaceAll(value, "&", "&amp;")

			values[i] = value
		}
		query[key] = values
	}
	c.Request().URL.RawQuery = query.Encode()
}

// sanitizeHeaders : 헤더 정화
func sanitizeHeaders(c echo.Context) {
	// User-Agent 정화
	userAgent := c.Request().Header.Get("User-Agent")
	if userAgent != "" {
		userAgent = strings.ReplaceAll(userAgent, "\r", "")
		userAgent = strings.ReplaceAll(userAgent, "\n", "")
		c.Request().Header.Set("User-Agent", userAgent)
	}
}
