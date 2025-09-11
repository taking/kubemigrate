package config

// NewMapConfigConverter 새로운 변환기 생성
func NewMapConfigConverter(data map[string]interface{}) *MapConfigConverter {
	return &MapConfigConverter{data: data}
}

// ToKubernetesConfig Kubernetes 설정으로 변환
func (c *MapConfigConverter) ToKubernetesConfig() *KubeConfig {
	if c.data == nil {
		return nil
	}

	config := &KubeConfig{}
	if val, ok := c.data["kubeconfig"].(string); ok {
		config.Config = val
	}
	if val, ok := c.data["namespace"].(string); ok {
		config.Namespace = val
	}
	return config
}

// ToMinioConfig MinIO 설정으로 변환
func (c *MapConfigConverter) ToMinioConfig() *MinioConfig {
	if c.data == nil {
		return nil
	}

	config := &MinioConfig{}
	if val, ok := c.data["endpoint"].(string); ok {
		config.Endpoint = val
	}
	if val, ok := c.data["accessKey"].(string); ok {
		config.AccessKey = val
	}
	if val, ok := c.data["secretKey"].(string); ok {
		config.SecretKey = val
	}
	if val, ok := c.data["useSSL"].(bool); ok {
		config.UseSSL = val
	}
	return config
}

// ToHelmConfig Helm 설정으로 변환
func (c *MapConfigConverter) ToHelmConfig() *KubeConfig {
	if c.data == nil {
		return nil
	}

	config := &KubeConfig{}
	if val, ok := c.data["kubeconfig"].(string); ok {
		config.Config = val
	}
	if val, ok := c.data["namespace"].(string); ok {
		config.Namespace = val
	}
	return config
}

// ToVeleroConfig Velero 설정으로 변환
func (c *MapConfigConverter) ToVeleroConfig() *VeleroConfig {
	if c.data == nil {
		return nil
	}

	config := &VeleroConfig{}
	if val, ok := c.data["kubeconfig"].(string); ok {
		config.KubeConfig = KubeConfig{Config: val}
	}

	// MinIO 설정 변환
	if minioData, ok := c.data["minio"].(map[string]interface{}); ok {
		minioConverter := &MapConfigConverter{data: minioData}
		if minioConfig := minioConverter.ToMinioConfig(); minioConfig != nil {
			config.MinioConfig = *minioConfig
		}
	}
	return config
}

// ToCacheConfig 캐시 설정으로 변환
func (c *MapConfigConverter) ToCacheConfig() *CacheConfig {
	if c.data == nil {
		return nil
	}

	config := &CacheConfig{}
	if val, ok := c.data["api_type"].(string); ok {
		config.ApiType = val
	}
	if val, ok := c.data["data"]; ok {
		config.Data = val
	}
	return config
}
