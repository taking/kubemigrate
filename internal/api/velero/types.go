package velero

import (
	"fmt"
	"time"
)

// InstallResult : 설치 결과
type InstallResult struct {
	Status           string                 `json:"status"`
	Message          string                 `json:"message"`
	VeleroNamespace  string                 `json:"velero_namespace"`
	MinioConnected   bool                   `json:"minio_connected"`
	BackupLocation   string                 `json:"backup_location"`
	InstallationTime time.Duration          `json:"installation_time"`
	Force            bool                   `json:"force"`
	Details          map[string]interface{} `json:"details,omitempty"`
}

// VeleroStatus : Velero 상태 정보
type VeleroStatus struct {
	PodsInstalled    bool   `json:"pods_installed"`
	HelmRelease      bool   `json:"helm_release"`
	ReleaseNamespace string `json:"release_namespace"`
	IsHealthy        bool   `json:"is_healthy"`
	ErrorMessage     string `json:"error_message,omitempty"`
}

// InstallationError : 설치 에러 정보
type InstallationError struct {
	Type        string   `json:"type"` // "helm_conflict", "pod_failed", "timeout"
	Message     string   `json:"message"`
	Details     string   `json:"details"`
	Suggestions []string `json:"suggestions"`
	Commands    []string `json:"commands"`
}

func (e *InstallationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Message, e.Details)
}

// InstallationProgress : 설치 진행 상황
type InstallationProgress struct {
	Step      string `json:"step"`
	Progress  int    `json:"progress"` // 0-100
	Message   string `json:"message"`
	Estimated string `json:"estimated"` // 남은 시간
	CanCancel bool   `json:"can_cancel"`
}

// BSLStatusInfo : BSL 상태 정보
type BSLStatusInfo struct {
	Phase              string   `json:"phase"`
	Message            string   `json:"message"`
	LastValidationTime string   `json:"last_validation_time"`
	MinioEndpoint      string   `json:"minio_endpoint"`
	Bucket             string   `json:"bucket"`
	Region             string   `json:"region"`
	SuggestedActions   []string `json:"suggested_actions"`
}

// VeleroResources : 네임스페이스 내 Velero 리소스 정보
type VeleroResources struct {
	HasResources   bool     `json:"has_resources"`
	HasPods        bool     `json:"has_pods"`
	HasSecrets     bool     `json:"has_secrets"`
	HasConfigMaps  bool     `json:"has_configmaps"`
	HasDeployments bool     `json:"has_deployments"`
	ResourceCount  int      `json:"resource_count"`
	ResourceTypes  []string `json:"resource_types"`
}
