package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	config := Load()

	if config == nil {
		t.Fatal("Load() returned nil")
	}

	// 기본값 확인
	if config.Server.Port != "9091" {
		t.Errorf("Expected default port 9091, got %s", config.Server.Port)
	}

	if config.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", config.Logging.Level)
	}

	if config.Logging.Format != "json" {
		t.Errorf("Expected default log format 'json', got %s", config.Logging.Format)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// 환경변수 설정
	os.Setenv("PORT", "8080")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "text")

	config := Load()

	if config.Server.Port != "8080" {
		t.Errorf("Expected port 8080 from env var, got %s", config.Server.Port)
	}

	if config.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug' from env var, got %s", config.Logging.Level)
	}

	if config.Logging.Format != "text" {
		t.Errorf("Expected log format 'text' from env var, got %s", config.Logging.Format)
	}

	// 환경변수 정리
	os.Unsetenv("PORT")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_FORMAT")
}

func TestGetDurationOrDefault(t *testing.T) {
	// 기본값 테스트
	duration := getDurationOrDefault("NON_EXISTENT_VAR", 30*time.Second)
	if duration != 30*time.Second {
		t.Errorf("Expected 30s default, got %v", duration)
	}

	// 환경변수 테스트
	os.Setenv("TEST_DURATION", "60s")
	duration = getDurationOrDefault("TEST_DURATION", 30*time.Second)
	if duration != 60*time.Second {
		t.Errorf("Expected 60s from env var, got %v", duration)
	}

	// 잘못된 형식 테스트
	os.Setenv("TEST_DURATION", "invalid")
	duration = getDurationOrDefault("TEST_DURATION", 30*time.Second)
	if duration != 30*time.Second {
		t.Errorf("Expected 30s default for invalid format, got %v", duration)
	}

	os.Unsetenv("TEST_DURATION")
}
