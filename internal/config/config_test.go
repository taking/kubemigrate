package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	config := Load()

	if config == nil {
		t.Fatal("Load() returned nil")
	}

	// 기본값 확인
	if config.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", config.Server.Port)
	}

	if config.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", config.Logging.Level)
	}

	if config.Logging.Format != "pretty" {
		t.Errorf("Expected default log format 'pretty', got %s", config.Logging.Format)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// 환경변수 설정
	if err := os.Setenv("SERVER_PORT", "8080"); err != nil {
		t.Fatalf("Failed to set SERVER_PORT env var: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
		t.Fatalf("Failed to set LOG_LEVEL env var: %v", err)
	}
	if err := os.Setenv("LOG_FORMAT", "text"); err != nil {
		t.Fatalf("Failed to set LOG_FORMAT env var: %v", err)
	}

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
	if err := os.Unsetenv("SERVER_PORT"); err != nil {
		t.Errorf("Failed to unset SERVER_PORT env var: %v", err)
	}
	if err := os.Unsetenv("LOG_LEVEL"); err != nil {
		t.Errorf("Failed to unset LOG_LEVEL env var: %v", err)
	}
	if err := os.Unsetenv("LOG_FORMAT"); err != nil {
		t.Errorf("Failed to unset LOG_FORMAT env var: %v", err)
	}
}

func TestGetDurationOrDefault(t *testing.T) {
	// 이 테스트는 pkg/config로 이동했으므로 제거
	// 필요시 pkg/config에서 테스트하도록 수정
}
