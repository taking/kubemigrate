// Package config 애플리케이션 설정을 관리합니다.
package config

import (
	"os"
	"time"
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
// 환경변수가 없으면 기본값 사용
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnvOrDefault("SERVER_HOST", DefaultServerHost),
			Port:         getEnvOrDefault("SERVER_PORT", DefaultServerPort),
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
