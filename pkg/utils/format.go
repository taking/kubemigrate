// Package utils 포맷팅 및 변환 유틸리티를 제공합니다.
package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// FormatBytes : 바이트 수를 읽기 쉬운 형태로 포맷팅
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GenerateCacheKey : 입력 문자열의 SHA256 해시 생성
func GenerateCacheKey(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// GenerateCompositeCacheKey : 여러 설정을 조합한 복합 캐시 키 생성
func GenerateCompositeCacheKey(configs ...string) string {
	// 모든 설정을 구분자로 연결
	combined := strings.Join(configs, "|")
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// FormatDuration : 지속 시간을 읽기 쉬운 형태로 포맷팅
func FormatDuration(duration int64) string {
	if duration < 1000 {
		return fmt.Sprintf("%d ns", duration)
	}
	if duration < 1000000 {
		return fmt.Sprintf("%.2f μs", float64(duration)/1000)
	}
	if duration < 1000000000 {
		return fmt.Sprintf("%.2f ms", float64(duration)/1000000)
	}
	return fmt.Sprintf("%.2f s", float64(duration)/1000000000)
}

// FormatPercentage : 백분율을 읽기 쉬운 형태로 포맷팅
func FormatPercentage(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}

// TruncateString : 문자열을 지정된 길이로 자름
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
