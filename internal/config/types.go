package config

import "time"

// === 애플리케이션 설정 ===

// Config 전체 애플리케이션 설정 구조체
type Config struct {
	Server   ServerConfig  // 서버 관련 설정
	Timeouts TimeoutConfig // 타임아웃 관련 설정
	Logging  LoggingConfig // 로깅 관련 설정
}

// ServerConfig 서버 호스트, 포트 및 타임아웃 설정
type ServerConfig struct {
	Host         string        // 서버 호스트
	Port         string        // 서버 포트
	ReadTimeout  time.Duration // 읽기 요청 제한 시간
	WriteTimeout time.Duration // 쓰기 요청 제한 시간
	IdleTimeout  time.Duration // 유휴 연결 제한 시간
}

// TimeoutConfig 각종 요청 및 헬스체크 타임아웃 설정
type TimeoutConfig struct {
	HealthCheck time.Duration // 헬스체크 타임아웃
	Request     time.Duration // 일반 요청 타임아웃
}

// LoggingConfig 로그 레벨 및 포맷 설정
type LoggingConfig struct {
	Level  string // 로그 레벨 (예: info, debug, warn, error)
	Format string // 로그 포맷 (예: json, text)
}

// === 서비스별 설정 ===

// KubeConfig Kubernetes 설정 구조체 (API용)
type KubeConfig struct {
	Config    string `json:"kubeconfig" binding:"required" example:"base64 인코딩된 KubeConfig 값"`
	Namespace string `json:"namespace,omitempty" example:"all" swaggerignore:"true"`
}

// MinioConfig MinIO 설정 구조체 (API용)
type MinioConfig struct {
	Endpoint  string `json:"endpoint" binding:"required" example:"127.0.0.1:9000"`
	AccessKey string `json:"accessKey" binding:"required" example:"your_minio_accessKey"`
	SecretKey string `json:"secretKey" binding:"required" example:"your_minio_secretKey"`
	UseSSL    bool   `json:"useSSL" example:"false"`
}

// VeleroConfig Velero 설정 구조체 (API용)
type VeleroConfig struct {
	KubeConfig  KubeConfig  `json:"kubeconfig" binding:"required"`
	MinioConfig MinioConfig `json:"minio" binding:"required"`
}

// === 플러그인 설정 (별칭으로 통일) ===

// KubernetesConfig Kubernetes 플러그인 설정 (KubeConfig 별칭)
type KubernetesConfig = KubeConfig

// HelmConfig Helm 플러그인 설정 (KubeConfig 별칭)
type HelmConfig = KubeConfig

// CacheConfig 캐시 설정
type CacheConfig struct {
	ApiType string `json:"api_type"`
	Data    any    `json:"data"`
}

// === 통합 플러그인 설정 ===

// PluginConfigData 통합 플러그인 설정
type PluginConfigData struct {
	Kubernetes *KubeConfig   `json:"kubernetes,omitempty"`
	Minio      *MinioConfig  `json:"minio,omitempty"`
	Helm       *KubeConfig   `json:"helm,omitempty"`
	Velero     *VeleroConfig `json:"velero,omitempty"`
	Cache      *CacheConfig  `json:"cache,omitempty"`
}

// === 설정 변환기 인터페이스 ===

// ConfigConverter 설정 변환 인터페이스
type ConfigConverter interface {
	ToKubernetesConfig() *KubeConfig
	ToMinioConfig() *MinioConfig
	ToHelmConfig() *KubeConfig
	ToVeleroConfig() *VeleroConfig
	ToCacheConfig() *CacheConfig
}

// MapConfigConverter map[string]interface{} 기반 변환기
type MapConfigConverter struct {
	data map[string]interface{}
}
