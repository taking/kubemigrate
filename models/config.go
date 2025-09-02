package models

type KubeConfig struct {
	KubeConfig string `json:"kubeconfig"` // Raw KubeConfig
	Namespace  string `json:"namespace,omitempty"`
	ChartName  string `json:"chartName,omitempty"` // helm
}

type MinioConfig struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	UseSSL    bool   `json:"useSSL"`
}
