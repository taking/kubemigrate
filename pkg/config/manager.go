package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// ConfigManager : 통합 설정 관리자
type ConfigManager struct {
	config *Config
}

// NewConfigManager : 새로운 설정 관리자 생성
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		config: Load(),
	}
}

// GetConfig : 전체 설정 반환
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// GetServerConfig : 서버 설정 반환
func (cm *ConfigManager) GetServerConfig() ServerConfig {
	return cm.config.Server
}

// GetTimeoutConfig : 타임아웃 설정 반환
func (cm *ConfigManager) GetTimeoutConfig() TimeoutConfig {
	return cm.config.Timeouts
}

// GetLoggingConfig : 로깅 설정 반환
func (cm *ConfigManager) GetLoggingConfig() LoggingConfig {
	return cm.config.Logging
}

// Reload : 설정 재로드
func (cm *ConfigManager) Reload() error {
	cm.config = Load()
	return nil
}

// ValidateConfig : 설정 검증
func (cm *ConfigManager) ValidateConfig() error {
	// 서버 설정 검증
	if err := cm.validateServerConfig(); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}

	// 타임아웃 설정 검증
	if err := cm.validateTimeoutConfig(); err != nil {
		return fmt.Errorf("timeout config validation failed: %w", err)
	}

	// 로깅 설정 검증
	if err := cm.validateLoggingConfig(); err != nil {
		return fmt.Errorf("logging config validation failed: %w", err)
	}

	return nil
}

// validateServerConfig : 서버 설정 검증
func (cm *ConfigManager) validateServerConfig() error {
	server := cm.config.Server

	if server.Host == "" {
		return fmt.Errorf("server host is required")
	}

	if server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if server.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}

	if server.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}

	if server.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive")
	}

	return nil
}

// validateTimeoutConfig : 타임아웃 설정 검증
func (cm *ConfigManager) validateTimeoutConfig() error {
	timeouts := cm.config.Timeouts

	if timeouts.HealthCheck <= 0 {
		return fmt.Errorf("health check timeout must be positive")
	}

	if timeouts.Request <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}

	return nil
}

// validateLoggingConfig : 로깅 설정 검증
func (cm *ConfigManager) validateLoggingConfig() error {
	logging := cm.config.Logging

	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[logging.Level] {
		return fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", logging.Level)
	}

	validFormats := map[string]bool{
		"json":   true,
		"text":   true,
		"pretty": true,
	}

	if !validFormats[logging.Format] {
		return fmt.Errorf("invalid log format: %s (valid: json, text, pretty)", logging.Format)
	}

	return nil
}

// GetEnvOrDefault : 환경변수 값 반환 (기본값 포함)
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDurationOrDefault : 환경변수 값을 Duration으로 변환 (기본값 포함)
func GetDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetIntOrDefault : 환경변수 값을 int로 변환 (기본값 포함)
func GetIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil && intValue > 0 {
			return intValue
		}
	}
	return defaultValue
}

// GetBoolOrDefault : 환경변수 값을 bool로 변환 (기본값 포함)
func GetBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		var boolValue bool
		if _, err := fmt.Sscanf(value, "%t", &boolValue); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// LoadEnvFile : .env 파일 로드
func LoadEnvFile() error {
	return godotenv.Load()
}
