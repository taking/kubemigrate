package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Timeouts TimeoutConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type TimeoutConfig struct {
	HealthCheck time.Duration
	Request     time.Duration
}

type LoggingConfig struct {
	Level  string
	Format string
}

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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
