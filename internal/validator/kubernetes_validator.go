package validator

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/taking/kubemigrate/internal/config"
)

// KubernetesValidator : KubernetesValidator 구조체
type KubernetesValidator struct {
	kubeconfigPattern *regexp.Regexp // kubeconfig 패턴
	namespacePattern  *regexp.Regexp // namespace 패턴
}

// NewKubernetesValidator : KubernetesValidator 초기화
func NewKubernetesValidator() *KubernetesValidator {
	return &KubernetesValidator{
		kubeconfigPattern: regexp.MustCompile(`apiVersion:\s*v1`),
		namespacePattern:  regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`),
	}
}

// ValidateKubernetesConfig : KubernetesConfig 검증
func (v *KubernetesValidator) ValidateKubernetesConfig(req *config.KubeConfig) (string, error) {
	if req.KubeConfig == "" {
		return "", fmt.Errorf("kubeconfig is required")
	}

	if len(req.KubeConfig) > 100000 { // 100KB limit
		return "", fmt.Errorf("kubeconfig too large (max 100KB)")
	}

	decodeSourceKubeConfig, _ := decodeIfBase64(req.KubeConfig)

	if !strings.Contains(decodeSourceKubeConfig, "apiVersion") {
		return "", fmt.Errorf("kubeconfig appears to be invalid (missing apiVersion)")
	}

	if req.Namespace != "" {
		if !v.isValidNamespace(req.Namespace) {
			return "", fmt.Errorf("invalid namespace format: must be lowercase alphanumeric with hyphens")
		}
	}

	return decodeSourceKubeConfig, nil
}

// DecodeIfBase64 : 클러스터의 KubeConfig을 Decode (공통 유틸리티 함수)
func DecodeIfBase64(s string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		// 실패하면 base64 아님, 원본 문자열 반환
		return s, nil
	}
	return string(decoded), nil
}

// decodeIfBase64 : 내부 헬퍼 함수 (DecodeIfBase64와 동일)
func decodeIfBase64(s string) (string, error) {
	return DecodeIfBase64(s)
}

// isValidNamespace : 네임스페이스 검증
func (v *KubernetesValidator) isValidNamespace(namespace string) bool {
	if len(namespace) == 0 || len(namespace) > 63 {
		return false
	}

	return v.namespacePattern.MatchString(namespace)
}
