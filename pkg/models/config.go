package models

type KubeConfig struct {
	KubeConfig string `json:"kubeconfig"`          // 원본 KubeConfig 데이터
	Namespace  string `json:"namespace,omitempty"` // 네임스페이스
	ChartName  string `json:"chartName,omitempty"` // 헬름 차트명
}

type MinioConfig struct {
	Endpoint   string `json:"endpoint"`             // MinIO 엔드포인트
	AccessKey  string `json:"accessKey"`            // 액세스 키
	SecretKey  string `json:"secretKey"`            // 시크릿 키
	UseSSL     bool   `json:"useSSL"`               // SSL 사용 여부
	BucketName string `json:"bucketName,omitempty"` // 버킷명
	Region     string `json:"region,omitempty"`     // 리전
}

type VeleroConfig struct {
	MinioConfig `json:"minio"`      // MinIO 설정
	KubeConfig  `json:"kubernetes"` // Kubernetes 설정
}

type HelmChartRequest struct {
	KubeConfig `json:"kubernetes"` // Kubernetes 설정
	ChartName  string              `json:"chartName"` // 헬름 차트명
}
