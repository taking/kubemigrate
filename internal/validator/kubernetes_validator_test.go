package validator

import (
	"encoding/base64"
	"testing"

	"github.com/taking/kubemigrate/pkg/config"
)

// TestNewKubernetesValidator : KubernetesValidator 생성자 테스트
// 새로 생성된 validator가 올바르게 초기화되었는지 확인
func TestNewKubernetesValidator(t *testing.T) {
	v := NewKubernetesValidator()
	if v == nil {
		t.Fatal("NewKubernetesValidator() returned nil")
	}
	if v.kubeconfigPattern == nil {
		t.Error("kubeconfigPattern is nil")
	}
	if v.namespacePattern == nil {
		t.Error("namespacePattern is nil")
	}
}

// TestKubernetesValidator_ValidateKubernetesConfig - Kubernetes 설정 검증 테스트
// 다양한 kubeconfig와 namespace 조합에 대한 검증 로직 테스트
func TestKubernetesValidator_ValidateKubernetesConfig(t *testing.T) {
	v := NewKubernetesValidator()

	tests := []struct {
		name    string
		config  *config.KubeConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "네임스페이스가 있는 유효한 설정",
			config: &config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
				Namespace:  "default",
			},
			wantErr: false,
		},
		{
			name: "네임스페이스가 없는 유효한 설정",
			config: &config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
				Namespace:  "",
			},
			wantErr: false,
		},
		{
			name: "Base64로 인코딩된 유효한 설정",
			config: &config.KubeConfig{
				KubeConfig: base64.StdEncoding.EncodeToString([]byte("apiVersion: v1\nkind: Config")),
				Namespace:  "kube-system",
			},
			wantErr: false,
		},
		{
			name: "빈 kubeconfig",
			config: &config.KubeConfig{
				KubeConfig: "",
				Namespace:  "default",
			},
			wantErr: true,
			errMsg:  "kubeconfig is required",
		},
		{
			name: "kubeconfig 크기 초과",
			config: &config.KubeConfig{
				KubeConfig: string(make([]byte, 100001)), // 100KB + 1 byte
				Namespace:  "default",
			},
			wantErr: true,
			errMsg:  "kubeconfig too large (max 100KB)",
		},
		{
			name: "apiVersion이 누락된 잘못된 kubeconfig",
			config: &config.KubeConfig{
				KubeConfig: "kind: Config",
				Namespace:  "default",
			},
			wantErr: true,
			errMsg:  "kubeconfig appears to be invalid (missing apiVersion)",
		},
		{
			name: "잘못된 네임스페이스 형식",
			config: &config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
				Namespace:  "INVALID_NAMESPACE",
			},
			wantErr: true,
			errMsg:  "invalid namespace format: must be lowercase alphanumeric with hyphens",
		},
		{
			name: "대문자가 포함된 네임스페이스",
			config: &config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
				Namespace:  "Default",
			},
			wantErr: true,
			errMsg:  "invalid namespace format: must be lowercase alphanumeric with hyphens",
		},
		{
			name: "하이픈으로 시작하는 네임스페이스",
			config: &config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
				Namespace:  "-invalid",
			},
			wantErr: true,
			errMsg:  "invalid namespace format: must be lowercase alphanumeric with hyphens",
		},
		{
			name: "하이픈으로 끝나는 네임스페이스",
			config: &config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
				Namespace:  "invalid-",
			},
			wantErr: true,
			errMsg:  "invalid namespace format: must be lowercase alphanumeric with hyphens",
		},
		{
			name: "숫자와 하이픈이 포함된 유효한 네임스페이스",
			config: &config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
				Namespace:  "test-namespace-123",
			},
			wantErr: false,
		},
		{
			name: "너무 긴 네임스페이스",
			config: &config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
				Namespace:  string(make([]byte, 64)), // 64 characters
			},
			wantErr: true,
			errMsg:  "invalid namespace format: must be lowercase alphanumeric with hyphens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := v.ValidateKubernetesConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateKubernetesConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != tt.errMsg {
					t.Errorf("ValidateKubernetesConfig() error message = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if decoded == "" {
					t.Error("ValidateKubernetesConfig() returned empty decoded config")
				}
			}
		})
	}
}

// TestDecodeIfBase64 - Base64 디코딩 유틸리티 함수 테스트
// Base64 인코딩된 문자열과 일반 문자열 모두 올바르게 처리하는지 확인
func TestDecodeIfBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Base64로 인코딩된 문자열",
			input:    base64.StdEncoding.EncodeToString([]byte("apiVersion: v1")),
			expected: "apiVersion: v1",
		},
		{
			name:     "일반 텍스트 문자열",
			input:    "apiVersion: v1",
			expected: "apiVersion: v1",
		},
		{
			name:     "빈 문자열",
			input:    "",
			expected: "",
		},
		{
			name:     "잘못된 Base64",
			input:    "invalid-base64!@#",
			expected: "invalid-base64!@#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeIfBase64(tt.input)
			if err != nil {
				t.Errorf("DecodeIfBase64() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("DecodeIfBase64() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestKubernetesValidator_isValidNamespace - 네임스페이스 유효성 검사 테스트
// Kubernetes 네임스페이스 명명 규칙에 따른 유효성 검사 로직 테스트
func TestKubernetesValidator_isValidNamespace(t *testing.T) {
	v := NewKubernetesValidator()

	tests := []struct {
		name      string
		namespace string
		want      bool
	}{
		{
			name:      "유효한 네임스페이스",
			namespace: "default",
			want:      true,
		},
		{
			name:      "숫자가 포함된 유효한 네임스페이스",
			namespace: "test123",
			want:      true,
		},
		{
			name:      "하이픈이 포함된 유효한 네임스페이스",
			namespace: "test-namespace",
			want:      true,
		},
		{
			name:      "숫자와 하이픈이 포함된 유효한 네임스페이스",
			namespace: "test-namespace-123",
			want:      true,
		},
		{
			name:      "빈 네임스페이스",
			namespace: "",
			want:      false,
		},
		{
			name:      "대문자가 포함된 네임스페이스",
			namespace: "Default",
			want:      false,
		},
		{
			name:      "하이픈으로 시작하는 네임스페이스",
			namespace: "-invalid",
			want:      false,
		},
		{
			name:      "하이픈으로 끝나는 네임스페이스",
			namespace: "invalid-",
			want:      false,
		},
		{
			name:      "너무 긴 네임스페이스",
			namespace: string(make([]byte, 64)),
			want:      false,
		},
		{
			name:      "특수문자가 포함된 네임스페이스",
			namespace: "test@namespace",
			want:      false,
		},
		{
			name:      "공백이 포함된 네임스페이스",
			namespace: "test namespace",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.isValidNamespace(tt.namespace); got != tt.want {
				t.Errorf("isValidNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
