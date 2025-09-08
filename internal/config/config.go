package config

import (
	"os"
	"time"
)

// Load : 환경변수 기반으로 Config 구조체 생성
// 환경변수가 없으면 기본값 사용
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnvOrDefault("SERVER_HOST", "localhost"),
			Port:         getEnvOrDefault("SERVER_PORT", "9091"),
			ReadTimeout:  getDurationOrDefault("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationOrDefault("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationOrDefault("IDLE_TIMEOUT", 120*time.Second),
		},
		Timeouts: TimeoutConfig{
			HealthCheck: getDurationOrDefault("HEALTH_CHECK_TIMEOUT", 5*time.Second),
			Request:     getDurationOrDefault("REQUEST_TIMEOUT", 30*time.Second),
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
