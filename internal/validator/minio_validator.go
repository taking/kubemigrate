package validator

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/taking/kubemigrate/pkg/config"
)

// MinioValidator : MinioValidator 구조체
type MinioValidator struct {
	endpointPattern  *regexp.Regexp // endpoint 패턴
	accessKeyPattern *regexp.Regexp // accessKey 패턴
	secretKeyPattern *regexp.Regexp // secretKey 패턴
}

// NewMinioValidator : MinioValidator 초기화
func NewMinioValidator() *MinioValidator {
	return &MinioValidator{
		// endpoint: host:port or domain, port optional
		endpointPattern: regexp.MustCompile(`^([a-zA-Z0-9.-]+)(:[0-9]{1,5})?$`),
		// accessKey: 최소 3글자 이상 영숫자
		accessKeyPattern: regexp.MustCompile(`^[A-Za-z0-9]{3,}$`),
		// secretKey: 공백 제외 8자 이상
		secretKeyPattern: regexp.MustCompile(`^\S{8,}$`),
	}
}

// ValidateMinioConfig : MinioConfig 검증
func (v *MinioValidator) ValidateMinioConfig(cfg *config.MinioConfig) error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("minio endpoint is required")
	}

	// url.Parse 를 통해 기본적인 포맷 확인
	if _, err := url.ParseRequestURI("http://" + cfg.Endpoint); err != nil {
		return fmt.Errorf("invalid minio endpoint format: %w", err)
	}

	if !v.endpointPattern.MatchString(cfg.Endpoint) {
		return fmt.Errorf("minio endpoint must be a valid hostname or IP with optional port")
	}

	if cfg.AccessKey == "" || !v.accessKeyPattern.MatchString(cfg.AccessKey) {
		return fmt.Errorf("invalid minio access key format")
	}

	if cfg.SecretKey == "" || !v.secretKeyPattern.MatchString(cfg.SecretKey) {
		return fmt.Errorf("invalid minio secret key format (min 8 chars, no spaces)")
	}

	return nil
}
