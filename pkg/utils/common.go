package utils

import (
	"encoding/base64"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strconv"
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

// GetStringOrDefault : value가 비어있지 않으면 value를 반환, 비어있으면 def를 반환
func GetStringOrDefault(value, def string) string {
	if value == "" {
		return def
	}
	return value
}

// GetBoolOrDefault : value가 비어있지 않으면 value를 반환, 비어있으면 def를 반환
func GetBoolOrDefault(value bool, def bool) bool {
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

// StringToIntOrDefault : string을 int로 변환, 실패하면 기본값 반환
func StringToIntOrDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// StringToBoolOrDefault : string을 bool로 변환, 실패하면 기본값 반환
func StringToBoolOrDefault(s string, def bool) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return def
	}
	return b
}
