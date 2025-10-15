package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// TestSecurityMiddleware : 보안 미들웨어 테스트
func TestSecurityMiddleware(t *testing.T) {
	e := echo.New()

	// 기본 설정으로 보안 미들웨어 생성
	securityMiddleware := SecurityMiddleware(DefaultSecurityConfig())

	// 테스트 핸들러
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// 미들웨어 적용
	e.Use(securityMiddleware)

	tests := []struct {
		name            string
		method          string
		path            string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:           "Valid GET request",
			method:         "GET",
			path:           "/test",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"X-Content-Type-Options": "nosniff",
				"X-Frame-Options":        "DENY",
				"X-XSS-Protection":       "1; mode=block",
			},
		},
		{
			name:           "Invalid method",
			method:         "INVALID",
			path:           "/test",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Path injection attempt",
			method:         "GET",
			path:           "/test/../../../etc/passwd",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			// 핸들러 실행
			e.ServeHTTP(rec, req)

			// 상태 코드 확인
			if tt.expectedStatus == http.StatusOK {
				if rec.Code != http.StatusOK {
					t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
				}
			} else {
				if rec.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
				}
			}

			// 보안 헤더 확인
			if tt.expectedHeaders != nil {
				for header, expectedValue := range tt.expectedHeaders {
					actualValue := rec.Header().Get(header)
					if actualValue != expectedValue {
						t.Errorf("Header %s should be %s, got %s", header, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

// TestCORSHeaders : CORS 헤더 테스트
func TestCORSHeaders(t *testing.T) {
	e := echo.New()

	config := &SecurityConfig{
		EnableCORS:            true,
		EnableSecurityHeaders: false,
		AllowedOrigins:        []string{"http://localhost:3000", "https://example.com"},
	}

	securityMiddleware := SecurityMiddleware(config)
	e.Use(securityMiddleware)

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
	}{
		{
			name:           "Allowed origin",
			origin:         "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
		},
		{
			name:           "Another allowed origin",
			origin:         "https://example.com",
			expectedOrigin: "https://example.com",
		},
		{
			name:           "Disallowed origin",
			origin:         "https://malicious.com",
			expectedOrigin: "",
		},
		{
			name:           "No origin",
			origin:         "",
			expectedOrigin: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			// 핸들러 실행
			e.ServeHTTP(rec, req)

			// CORS 헤더 확인
			actualOrigin := rec.Header().Get("Access-Control-Allow-Origin")
			if actualOrigin != tt.expectedOrigin {
				t.Errorf("Expected origin %s, got %s", tt.expectedOrigin, actualOrigin)
			}

			// 다른 CORS 헤더들 확인
			expectedMethods := "GET, POST, PUT, DELETE, OPTIONS"
			actualMethods := rec.Header().Get("Access-Control-Allow-Methods")
			if actualMethods != expectedMethods {
				t.Errorf("Expected methods %s, got %s", expectedMethods, actualMethods)
			}

			expectedHeaders := "Content-Type, Authorization, X-Requested-With"
			actualHeaders := rec.Header().Get("Access-Control-Allow-Headers")
			if actualHeaders != expectedHeaders {
				t.Errorf("Expected headers %s, got %s", expectedHeaders, actualHeaders)
			}
		})
	}
}

// TestInputSanitization : 입력 정화 테스트
func TestInputSanitization(t *testing.T) {
	e := echo.New()

	sanitizationMiddleware := InputSanitizationMiddleware()
	e.Use(sanitizationMiddleware)

	e.GET("/test", func(c echo.Context) error {
		// 쿼리 파라미터 반환
		param := c.QueryParam("test")
		return c.String(http.StatusOK, param)
	})

	req := httptest.NewRequest("GET", "/test?test=<script>alert('xss')</script>", nil)
	rec := httptest.NewRecorder()

	// 핸들러 실행
	e.ServeHTTP(rec, req)

	// XSS 공격이 정화되었는지 확인
	response := rec.Body.String()
	if response == "<script>alert('xss')</script>" {
		t.Error("XSS attack was not sanitized")
	}
}

// TestRateLimitMiddleware : 요청 제한 미들웨어 테스트
func TestRateLimitMiddleware(t *testing.T) {
	e := echo.New()

	rateLimitMiddleware := RateLimitMiddleware(10) // 분당 10회
	e.Use(rateLimitMiddleware)

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// 핸들러 실행
	e.ServeHTTP(rec, req)

	// 현재는 기본 구현이므로 항상 성공
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// TestDefaultSecurityConfig : 기본 보안 설정 테스트
func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	if !config.EnableCORS {
		t.Error("EnableCORS should be true")
	}
	if !config.EnableSecurityHeaders {
		t.Error("EnableSecurityHeaders should be true")
	}
	found := false
	for _, origin := range config.AllowedOrigins {
		if origin == "*" {
			found = true
			break
		}
	}
	if !found {
		t.Error("AllowedOrigins should contain '*'")
	}
	if config.MaxRequestSize != int64(10*1024*1024) {
		t.Errorf("MaxRequestSize should be %d, got %d", int64(10*1024*1024), config.MaxRequestSize)
	}
}

// TestIsAllowedOrigin : Origin 허용 검증 테스트
func TestIsAllowedOrigin(t *testing.T) {
	allowedOrigins := []string{"http://localhost:3000", "https://example.com"}

	tests := []struct {
		origin   string
		expected bool
	}{
		{"http://localhost:3000", true},
		{"https://example.com", true},
		{"https://malicious.com", false},
		{"", false},
		{"http://localhost:3001", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			result := isAllowedOrigin(tt.origin, allowedOrigins)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsValidHTTPMethod : HTTP 메서드 검증 테스트
func TestIsValidHTTPMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected bool
	}{
		{"GET", true},
		{"POST", true},
		{"PUT", true},
		{"DELETE", true},
		{"OPTIONS", true},
		{"HEAD", true},
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := isValidHTTPMethod(tt.method)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsValidPath : 경로 검증 테스트
func TestIsValidPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/api/v1/test", true},
		{"/api/v1/../etc/passwd", false},
		{"/api/v1/", true},
		{"/" + string(make([]byte, 3000)), false}, // 너무 긴 경로
		{"/api/v1/<script>", false},
		{"/api/v1/test?param=value", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isValidPath(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
