// Package config 애플리케이션 설정을 관리합니다.
package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

// 서버 관련 상수
const (
	// 기본 서버 설정
	DefaultServerHost = "localhost"
	DefaultServerPort = "8080"

	// 타임아웃 설정
	DefaultReadTimeout  = 30 * time.Second
	DefaultWriteTimeout = 30 * time.Second
	DefaultIdleTimeout  = 120 * time.Second

	// 요청 타임아웃
	DefaultRequestTimeout = 30 * time.Second

	// 헬스체크 타임아웃
	DefaultHealthCheckTimeout = 5 * time.Second
)

// Load : 환경변수 기반으로 Config 구조체 생성
// .env 파일이 있으면 먼저 로드하고, 환경변수가 없으면 기본값 사용
func Load() *Config {
	// .env 파일 로드 (파일이 없어도 에러 무시)
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Host:         getEnvOrDefault("SERVER_HOST", DefaultServerHost),
			Port:         getPortOrDefault(),
			ReadTimeout:  getDurationOrDefault("READ_TIMEOUT", DefaultReadTimeout),
			WriteTimeout: getDurationOrDefault("WRITE_TIMEOUT", DefaultWriteTimeout),
			IdleTimeout:  getDurationOrDefault("IDLE_TIMEOUT", DefaultIdleTimeout),
		},
		Timeouts: TimeoutConfig{
			HealthCheck: getDurationOrDefault("HEALTH_CHECK_TIMEOUT", DefaultHealthCheckTimeout),
			Request:     getDurationOrDefault("REQUEST_TIMEOUT", DefaultRequestTimeout),
		},
		Logging: LoggingConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: getEnvOrDefault("LOG_FORMAT", "pretty"),
		},
	}
}

// getEnvOrDefault : 환경변수 key가 존재하면 반환, 없으면 defaultValue 반환
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDurationOrDefault : 환경변수 key가 존재하면 time.Duration로 변환하여 반환
// 변환 실패 시 기본값(defaultValue) 사용
func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getPortOrDefault : PORT 또는 SERVER_PORT 환경변수를 확인하여 포트 반환
// 둘 다 없으면 기본값 사용
func getPortOrDefault() string {
	// PORT 환경변수 우선 확인 (일반적인 관례)
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	// SERVER_PORT 환경변수 확인
	if port := os.Getenv("SERVER_PORT"); port != "" {
		return port
	}
	// 둘 다 없으면 기본값 사용
	return DefaultServerPort
}
