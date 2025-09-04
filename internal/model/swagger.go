package model

type KubeConfigRequest struct {
	KubeConfig string `json:"kubeconfig" binding:"required" example:"base64 인코딩된 KubeConfig 값"` // [필수] Base64 인코딩된 KubeConfig 값
	Namespace  string `json:"namespace,omitempty" example:"all"`                                // [옵션] 네임스페이스 명 (기본 값 : 'default', 전체 조회  : 'all')
}

type HelmConfigRequest struct {
	KubeConfig string `json:"kubeconfig" binding:"required" example:"base64 인코딩된 KubeConfig 값"` // [필수] Base64 인코딩된 KubeConfig 값
	ChartName  string `json:"chartName,omitempty" binding:"required" example:"podinfo"`         // [옵션] 헬름 차트 명
}

type MinioConfigRequest struct {
	UseSSL     bool   `json:"useSSL" binding:"required" example:"false"`                   // [필수] minio useSSL 여부 (false: http / true: https)
	Endpoint   string `json:"endpoint" binding:"required" example:"127.0.0.1:9000"`        // [필수] minio endpoint 주소
	AccessKey  string `json:"accessKey" binding:"required" example:"your_minio_accessKey"` // [필수] minio accessKey
	SecretKey  string `json:"secretKey" binding:"required" example:"your_minio_secretKey"` // [필수] minio secretKey
	BucketName string `json:"bucketName" example:"velero"`                                 // [옵션] minio Bucket  (기본값: velero)
	Region     string `json:"region" example:"us-east-1"`                                  // [옵션] minio Region (기본값 : us-east-1)
}

type VeleroConfigRequest struct {
	MinioConfigRequest `json:"minio" binding:"required"`
	KubeConfig         string `json:"kubeconfig" binding:"required" example:"base64 인코딩된 KubeConfig 값"` // [필수] Base64 인코딩된 KubeConfig 값
}

type HelmInstallChartRequest struct {
	HelmConfigRequest
	Namespace string `json:"namespace,omitempty" example:"all"` // [옵션] 네임스페이스 명 (기본 값 : 'default', 전체 조회  : 'all')
}

type SwaggerSuccessResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}

type SwaggerErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type SwaggerStatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
