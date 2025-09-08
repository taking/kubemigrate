package config

import "time"

// Config : 전체 애플리케이션 설정 구조체
type Config struct {
	Server   ServerConfig  // 서버 관련 설정
	Timeouts TimeoutConfig // 타임아웃 관련 설정
	Logging  LoggingConfig // 로깅 관련 설정
}

// ServerConfig : 서버 포트 및 타임아웃 설정
type ServerConfig struct {
	Port         string        // 서버 포트
	ReadTimeout  time.Duration // 읽기 요청 제한 시간
	WriteTimeout time.Duration // 쓰기 요청 제한 시간
	IdleTimeout  time.Duration // 유휴 연결 제한 시간
}

// TimeoutConfig : 각종 요청 및 헬스체크 타임아웃 설정
type TimeoutConfig struct {
	HealthCheck time.Duration // 헬스체크 타임아웃
	Request     time.Duration // 일반 요청 타임아웃
}

// LoggingConfig : 로그 레벨 및 포맷 설정
type LoggingConfig struct {
	Level  string // 로그 레벨 (예: info, debug, warn, error)
	Format string // 로그 포맷 (예: json, text)
}

// KubeConfig Kubernetes 설정 구조체
type KubeConfig struct {
	KubeConfig string `json:"kubeconfig" binding:"required"`
	Namespace  string `json:"namespace,omitempty"`
}

// MinioConfig MinIO 설정 구조체
type MinioConfig struct {
	Endpoint        string `json:"endpoint" binding:"required"`
	AccessKeyID     string `json:"accessKeyID" binding:"required"`
	SecretAccessKey string `json:"secretAccessKey" binding:"required"`
	UseSSL          bool   `json:"useSSL"`
}

// HelmConfig Helm 설정 구조체
type HelmConfig struct {
	KubeConfig KubeConfig `json:"kubeconfig" binding:"required"`
}

// VeleroConfig Velero 설정 구조체
type VeleroConfig struct {
	KubeConfig  KubeConfig  `json:"kubeconfig" binding:"required"`
	MinioConfig MinioConfig `json:"minio" binding:"required"`
}
