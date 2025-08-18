package utils

import (
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ParseRestConfigFromRaw converts a raw kubeconfig YAML string into a *rest.Config.
func ParseRestConfigFromRaw(rawKubeconfig string) (*rest.Config, error) {
	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(rawKubeconfig))
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}
	return cfg, nil
}
