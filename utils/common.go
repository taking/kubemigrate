package utils

import (
	"encoding/base64"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DecodeIfBase64 클러스터의 KubeConfig을 Decode
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

// DefaultBool : value가 비어 있지 않으면 false를 반환하고, 비어 있으면 true(기본값)를 반환합니다.
func DefaultBool(value bool, def bool) bool {
	if &value == nil {
		return def
	}
	return value
}
