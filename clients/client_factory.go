package clients

import (
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"taking.kr/velero/models"
)

// ClientFactory provides unified client creation
type ClientFactory struct{}

func NewClientFactory() *ClientFactory {
	return &ClientFactory{}
}

// CreateRESTConfig creates REST config from KubeConfig
func (f *ClientFactory) CreateRESTConfig(cfg models.KubeConfig) (*rest.Config, error) {
	var restCfg *rest.Config
	var err error

	if cfg.KubeConfig != "" {
		restCfg, err = clientcmd.RESTConfigFromKubeConfig([]byte(cfg.KubeConfig))
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
		}
	} else {
		restCfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
		}
	}

	return restCfg, nil
}

// ResolveNamespace :  resolves namespace with defaults
func (f *ClientFactory) ResolveNamespace(cfg *models.KubeConfig, defaultNS string) string {
	if cfg.Namespace != "" {
		return cfg.Namespace
	}
	return defaultNS
}
