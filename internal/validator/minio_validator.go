package validator

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/taking/kubemigrate/internal/config"
)

// MinioValidator : MinioValidator 구조체
type MinioValidator struct {
	hostPortPattern  *regexp.Regexp // host:port 패턴 (IPv4, IPv6, 도메인)
	accessKeyPattern *regexp.Regexp // accessKey 패턴
	secretKeyPattern *regexp.Regexp // secretKey 패턴
}

// NewMinioValidator : MinioValidator 초기화
func NewMinioValidator() *MinioValidator {
	return &MinioValidator{
		// host:port 패턴 - IPv6, IPv4, 도메인 모두 지원
		hostPortPattern: regexp.MustCompile(`^(\[([0-9a-fA-F]{1,4}:){2,7}[0-9a-fA-F]{1,4}\]|([0-9]{1,3}\.){3}[0-9]{1,3}|[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*)(:[0-9]{1,5})?$`),
		// accessKey: 최소 3글자 이상 영숫자
		accessKeyPattern: regexp.MustCompile(`^[A-Za-z0-9]{3,}$`),
		// secretKey: 공백 제외 8자 이상
		secretKeyPattern: regexp.MustCompile(`^\S{8,}$`),
	}
}

// ValidateMinioConfig : MinioConfig 검증
func (v *MinioValidator) ValidateMinioConfig(cfg *config.MinioConfig) error {
	// 1. Endpoint 검증
	if err := v.validateEndpoint(cfg.Endpoint, cfg.UseSSL); err != nil {
		return err
	}

	// 2. AccessKey 검증
	if err := v.validateAccessKey(cfg.AccessKey); err != nil {
		return err
	}

	// 3. SecretKey 검증
	if err := v.validateSecretKey(cfg.SecretKey); err != nil {
		return err
	}

	return nil
}

// validateEndpoint : Endpoint 검증 (전체 URL 또는 host:port 형식 지원)
func (v *MinioValidator) validateEndpoint(endpoint string, useSSL bool) error {
	if endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	// 전체 URL 형식인지 확인 (http:// 또는 https://로 시작)
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return v.validateFullURL(endpoint, useSSL)
	}

	// host:port 형식 검증
	return v.validateHostPort(endpoint)
}

// validateFullURL : 전체 URL 형식 검증
func (v *MinioValidator) validateFullURL(endpoint string, useSSL bool) error {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid URL format: %s", err.Error())
	}

	// 프로토콜 검증
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("endpoint must use http:// or https:// protocol, got: %s", parsedURL.Scheme)
	}

	// SSL 설정과 프로토콜 일치성 검증
	if useSSL && parsedURL.Scheme != "https" {
		return fmt.Errorf("endpoint uses http:// but UseSSL is true - use https:// for secure connections")
	}
	if !useSSL && parsedURL.Scheme == "https" {
		return fmt.Errorf("endpoint uses https:// but UseSSL is false - set UseSSL to true for secure connections")
	}

	// 호스트 검증
	if parsedURL.Host == "" {
		return fmt.Errorf("endpoint must include a valid hostname or IP address")
	}

	// 호스트 부분만 추출하여 host:port 형식으로 검증
	host := parsedURL.Host
	if !v.hostPortPattern.MatchString(host) {
		return fmt.Errorf("endpoint host '%s' is not a valid hostname, IP address, or IPv6 address", host)
	}

	return nil
}

// validateHostPort : host:port 형식 검증
func (v *MinioValidator) validateHostPort(endpoint string) error {
	if !v.hostPortPattern.MatchString(endpoint) {
		return fmt.Errorf("endpoint '%s' must be in format 'hostname:port' or 'hostname' (e.g., 'localhost:9000', 'minio.example.com', '[::1]:9000')", endpoint)
	}

	// URL 파싱을 통한 추가 검증
	testURL := "http://" + endpoint
	if _, err := url.Parse(testURL); err != nil {
		return fmt.Errorf("endpoint '%s' is not a valid hostname or IP address: %s", endpoint, err.Error())
	}

	return nil
}

// validateAccessKey : AccessKey 검증
func (v *MinioValidator) validateAccessKey(accessKey string) error {
	if accessKey == "" {
		return fmt.Errorf("access key is required")
	}

	if !v.accessKeyPattern.MatchString(accessKey) {
		return fmt.Errorf("access key must be at least 3 characters long and contain only letters and numbers, got: '%s'", accessKey)
	}

	return nil
}

// validateSecretKey : SecretKey 검증
func (v *MinioValidator) validateSecretKey(secretKey string) error {
	if secretKey == "" {
		return fmt.Errorf("secret key is required")
	}

	if !v.secretKeyPattern.MatchString(secretKey) {
		return fmt.Errorf("secret key must be at least 8 characters long and cannot contain spaces, got: %d characters", len(secretKey))
	}

	return nil
}
