package utils

import (
	"context"
	"fmt"

	"github.com/taking/kubemigrate/pkg/client"
	v1 "k8s.io/api/core/v1"
)

// GetPods : Kubernetes pods 조회
func GetPods(client client.Client, ctx context.Context, namespace, name string) (*v1.PodList, error) {
	pods, err := client.Kubernetes().GetPods(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}
	return pods.(*v1.PodList), nil
}

// IsVeleroInstalled : Velero 설치 여부 확인
func IsVeleroInstalled(pods *v1.PodList) bool {
	for _, pod := range pods.Items {
		if pod.Labels["component"] == "velero" && pod.Labels["deploy"] == "velero" {
			return true
		}
	}
	return false
}

// CheckVeleroInstallation : Velero 설치 상태 확인
func CheckVeleroInstallation(client client.Client, ctx context.Context, namespace string) error {
	pods, err := GetPods(client, ctx, namespace, "")
	if err != nil {
		return err
	}

	if !IsVeleroInstalled(pods) {
		return fmt.Errorf("velero not found in namespace %s", namespace)
	}
	return nil
}
