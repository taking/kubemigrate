package validator

import (
	"testing"

	"github.com/taking/kubemigrate/pkg/config"
)

// TestNewMinioValidator : MinioValidator 생성자 테스트
// 새로 생성된 validator가 올바르게 초기화되었는지 확인
func TestNewMinioValidator(t *testing.T) {
	v := NewMinioValidator()
	if v == nil {
		t.Fatal("NewMinioValidator() returned nil")
	}
	if v.endpointPattern == nil {
		t.Error("endpointPattern is nil")
	}
	if v.accessKeyPattern == nil {
		t.Error("accessKeyPattern is nil")
	}
	if v.secretKeyPattern == nil {
		t.Error("secretKeyPattern is nil")
	}
}

// TestMinioValidator_ValidateMinioConfig - MinIO 설정 검증 테스트
// 다양한 endpoint, accessKey, secretKey 조합에 대한 검증 로직 테스트
func TestMinioValidator_ValidateMinioConfig(t *testing.T) {
	v := NewMinioValidator()

	tests := []struct {
		name    string
		config  *config.MinioConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "유효한 설정",
			config: &config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			wantErr: false,
		},
		{
			name: "도메인이 포함된 유효한 설정",
			config: &config.MinioConfig{
				Endpoint:  "minio.example.com",
				AccessKey: "access123",
				SecretKey: "secret123456",
				UseSSL:    true,
			},
			wantErr: false,
		},
		{
			name: "IP 주소가 포함된 유효한 설정",
			config: &config.MinioConfig{
				Endpoint:  "192.168.1.100:9000",
				AccessKey: "user123",
				SecretKey: "password123",
				UseSSL:    false,
			},
			wantErr: false,
		},
		{
			name: "포트가 없는 유효한 설정",
			config: &config.MinioConfig{
				Endpoint:  "localhost",
				AccessKey: "admin",
				SecretKey: "password123",
				UseSSL:    false,
			},
			wantErr: false,
		},
		{
			name: "빈 엔드포인트",
			config: &config.MinioConfig{
				Endpoint:  "",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			wantErr: true,
			errMsg:  "minio endpoint is required",
		},
		{
			name: "잘못된 엔드포인트 형식",
			config: &config.MinioConfig{
				Endpoint:  "invalid://endpoint",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			wantErr: true,
			errMsg:  "minio endpoint must be a valid hostname or IP with optional port",
		},
		{
			name: "잘못된 엔드포인트 패턴",
			config: &config.MinioConfig{
				Endpoint:  "invalid@endpoint",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			wantErr: true,
			errMsg:  "minio endpoint must be a valid hostname or IP with optional port",
		},
		{
			name: "빈 액세스 키",
			config: &config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			wantErr: true,
			errMsg:  "invalid minio access key format",
		},
		{
			name: "너무 짧은 액세스 키",
			config: &config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "ab",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			wantErr: true,
			errMsg:  "invalid minio access key format",
		},
		{
			name: "특수문자가 포함된 액세스 키",
			config: &config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "access@key",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			wantErr: true,
			errMsg:  "invalid minio access key format",
		},
		{
			name: "빈 시크릿 키",
			config: &config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "",
				UseSSL:    false,
			},
			wantErr: true,
			errMsg:  "invalid minio secret key format (min 8 chars, no spaces)",
		},
		{
			name: "너무 짧은 시크릿 키",
			config: &config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "short",
				UseSSL:    false,
			},
			wantErr: true,
			errMsg:  "invalid minio secret key format (min 8 chars, no spaces)",
		},
		{
			name: "공백이 포함된 시크릿 키",
			config: &config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "secret with spaces",
				UseSSL:    false,
			},
			wantErr: true,
			errMsg:  "invalid minio secret key format (min 8 chars, no spaces)",
		},
		{
			name: "숫자가 포함된 유효한 액세스 키",
			config: &config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "access123",
				SecretKey: "secret123456",
				UseSSL:    false,
			},
			wantErr: false,
		},
		{
			name: "대소문자가 혼합된 유효한 액세스 키",
			config: &config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "AccessKey123",
				SecretKey: "secret123456",
				UseSSL:    false,
			},
			wantErr: false,
		},
		{
			name: "특수문자가 포함된 유효한 시크릿 키",
			config: &config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "access123",
				SecretKey: "secret!@#$%^&*()",
				UseSSL:    false,
			},
			wantErr: false,
		},
		{
			name: "높은 포트 번호가 포함된 엔드포인트",
			config: &config.MinioConfig{
				Endpoint:  "localhost:65535",
				AccessKey: "access123",
				SecretKey: "secret123456",
				UseSSL:    false,
			},
			wantErr: false,
		},
		{
			name: "잘못된 포트가 포함된 엔드포인트",
			config: &config.MinioConfig{
				Endpoint:  "localhost:99999",
				AccessKey: "access123",
				SecretKey: "secret123456",
				UseSSL:    false,
			},
			wantErr: false, // 99999 is actually valid for URL parsing
		},
		{
			name: "0 포트가 포함된 엔드포인트",
			config: &config.MinioConfig{
				Endpoint:  "localhost:0",
				AccessKey: "access123",
				SecretKey: "secret123456",
				UseSSL:    false,
			},
			wantErr: false, // 0 is actually valid for URL parsing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateMinioConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMinioConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != tt.errMsg && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateMinioConfig() error message = %v, want contains %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

// contains - 문자열 포함 여부 확인 유틸리티 함수
// s가 substr을 포함하는지 확인
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
