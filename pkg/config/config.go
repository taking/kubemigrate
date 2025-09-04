package config

import (
	"os"
	"strconv"
	"time"
)

// Config : 전체 애플리케이션 설정 구조체
type Config struct {
	Server   ServerConfig  // 서버 관련 설정
	Timeouts TimeoutConfig // 타임아웃 관련 설정
	Logging  LoggingConfig // 로깅 관련 설정
}

// ServerConfig : 서버 포트 및 타임아웃 설정
type ServerConfig struct {
	Port         string        // 서버 포트
	ReadTimeout  time.Duration // 읽기 요청 제한 시간
	WriteTimeout time.Duration // 쓰기 요청 제한 시간
	IdleTimeout  time.Duration // 유휴 연결 제한 시간
}

// TimeoutConfig : 각종 요청 및 헬스체크 타임아웃 설정
type TimeoutConfig struct {
	HealthCheck time.Duration // 헬스체크 타임아웃
	Request     time.Duration // 일반 요청 타임아웃
}

// LoggingConfig : 로그 레벨 및 포맷 설정
type LoggingConfig struct {
	Level  string // 로그 레벨 (예: info, debug, warn, error)
	Format string // 로그 포맷 (예: json, text)
}

// Load : 환경변수 기반으로 Config 구조체 생성
// 환경변수가 없으면 기본값 사용
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvOrDefault("PORT", "9091"),
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
			Format: getEnvOrDefault("LOG_FORMAT", "json"),
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

// getIntOrDefault : 환경변수 key가 존재하면 int로 변환하여 반환
// 변환 실패 시 기본값(defaultValue) 사용
func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
