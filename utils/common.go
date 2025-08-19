package utils

import (
	"encoding/base64"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DecodeIfBase64 클러스터의 KubeConfig을 디코딩
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
