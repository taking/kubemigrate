package utils

import (
	"encoding/base64"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DecodeIfBase64 : 클러스터의 KubeConfig을 Decode
func DecodeIfBase64(s string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		// 실패하면 base64 아님
		return "", fmt.Errorf("not valid base64: %w", err)
	}
	return string(decoded), nil
}

// StripManagedFields 리소스의 metadata.managedFields를 제거
func StripManagedFields(obj metav1.Object) {
	obj.SetManagedFields(nil)
}

// DefaultString : value가 비어 있지 않으면 value를 반환하고, 비어 있으면 def(기본값)를 반환합니다.
func DefaultString(value, def string) string {
	if value == "" {
		return def
	}
	return value
}

// DefaultBool : value가 기본값이면 def를 반환하고, 그렇지 않으면 value를 반환합니다.
func DefaultBool(value bool, def bool) bool {
	if !value && def {
		return def
	}
	return value
}

// 간단한 파일 복사 함수
func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}
