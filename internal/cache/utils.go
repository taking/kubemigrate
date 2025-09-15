package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/taking/kubemigrate/pkg/config"
)

// maskMiddle : 문자열을 앞/뒤 n자만 남기고 가운데 마스킹 처리
func maskMiddle(s string, prefix, suffix int) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) <= prefix+suffix {
		return strings.Repeat("*", len(s))
	}
	return s[:prefix] + strings.Repeat("*", len(s)-prefix-suffix) + s[len(s)-suffix:]
}

// maskString : 기본 문자열 마스킹 (앞 4, 뒤 4)
func maskString(s string) string {
	return maskMiddle(s, 4, 4)
}

// maskKubeConfigString : kubeconfig 문자열 마스킹 (앞 6, 뒤 6)
func maskKubeConfigString(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) <= 12 {
		return strings.Repeat("*", len(s))
	}
	return s[:6] + "..." + s[len(s)-6:]
}

// maskKubeConfig : kubeconfig 마스킹 처리
func maskKubeConfig(kubeConfig config.KubeConfig) MaskedKubeConfig {
	return MaskedKubeConfig{
		KubeConfig: maskKubeConfigString(kubeConfig.KubeConfig),
		HasConfig:  kubeConfig.KubeConfig != "",
	}
}

// maskKubernetesConfig : Kubernetes 설정 마스킹 처리 (별칭)
func maskKubernetesConfig(kubeConfig config.KubeConfig) MaskedKubernetesConfig {
	return maskKubeConfig(kubeConfig)
}

// maskMinioConfig : MinIO 설정 마스킹 처리
func maskMinioConfig(minioConfig config.MinioConfig) MaskedMinioConfig {
	return MaskedMinioConfig{
		Endpoint:  minioConfig.Endpoint,
		AccessKey: maskString(minioConfig.AccessKey),
		SecretKey: maskString(minioConfig.SecretKey),
		UseSSL:    minioConfig.UseSSL,
		HasConfig: minioConfig.Endpoint != "",
	}
}

// maskVeleroConfig : Velero 설정 마스킹 처리
func maskVeleroConfig(veleroConfig config.VeleroConfig, minioConfig config.MinioConfig) MaskedVeleroConfig {
	return MaskedVeleroConfig{
		Kubernetes: maskKubernetesConfig(veleroConfig.KubeConfig),
		Minio:      maskMinioConfig(minioConfig),
		HasConfig:  veleroConfig.KubeConfig.KubeConfig != "" || minioConfig.Endpoint != "",
	}
}

// maskHelmConfig : Helm 설정 마스킹 처리
func maskHelmConfig(helmConfig config.KubeConfig) MaskedHelmConfig {
	return MaskedHelmConfig{
		KubeConfig: maskKubeConfigString(helmConfig.KubeConfig),
		HasConfig:  helmConfig.KubeConfig != "",
	}
}

// getApiTypeFromKey : 캐시 키에서 API 타입 추출
func getApiTypeFromKey(key string) string {
	if len(key) == 64 {
		return "kubernetes"
	}
	lower := strings.ToLower(key)
	switch {
	case strings.Contains(lower, "minio"):
		return "minio"
	case strings.Contains(lower, "velero"):
		return "velero"
	case strings.Contains(lower, "helm"):
		return "helm"
	default:
		return "kubernetes"
	}
}

// generateReadableKey : 읽기 쉬운 키 생성
func generateReadableKey(apiType, key string) string {
	if len(key) <= 12 {
		return apiType + ":" + key
	}
	return apiType + ":" + key[:6] + "..." + key[len(key)-6:]
}

// calculateAgeSeconds : 생성 시간으로부터 경과된 초 계산
func calculateAgeSeconds(createdAt time.Time) int {
	return int(time.Since(createdAt).Seconds())
}

// calculateRemainingSeconds : 남은 TTL 초 계산
func calculateRemainingSeconds(createdAt time.Time, ttl time.Duration) int {
	remaining := time.Until(createdAt.Add(ttl))
	if remaining < 0 {
		return 0
	}
	return int(remaining.Seconds())
}

// generateCacheKeyFromConfig : 설정으로부터 캐시 키 생성
func generateCacheKeyFromConfig(kubeConfig config.KubeConfig) string {
	hash := sha256.Sum256([]byte(kubeConfig.KubeConfig))
	return hex.EncodeToString(hash[:])
}
