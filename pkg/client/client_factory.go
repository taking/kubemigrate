package client

import (
	"fmt"

	"github.com/taking/kubemigrate/pkg/models"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientFactory : Kubernetes 클라이언트 생성을 통합 관리하는 팩토리
type ClientFactory struct{}

// NewClientFactory : ClientFactory 객체 생성
func NewClientFactory() *ClientFactory {
	return &ClientFactory{}
}

// CreateRESTConfig : KubeConfig로 RESTConfig 설정 생성
// cfg.KubeConfig가 지정되어 있으면 해당 kubeconfig를 파싱하여 RESTConfig 생성
// 지정되어 있지 않으면 InClusterConfig를 사용하여 클러스터 내부 설정으로 생성
func (f *ClientFactory) CreateRESTConfig(cfg models.KubeConfig) (*rest.Config, error) {
	var restCfg *rest.Config
	var err error

	if cfg.KubeConfig != "" {
		// 외부 kubeconfig 사용
		restCfg, err = clientcmd.RESTConfigFromKubeConfig([]byte(cfg.KubeConfig))
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
		}
	} else {
		// 클러스터 내부 설정 사용
		restCfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
		}
	}

	return restCfg, nil
}

// getNamespaceOrDefault : 네임스페이스가 설정되어 있으면 사용, 없으면 기본값 반환
func getNamespaceOrDefault(namespace, defaultNS string) string {
	if namespace != "" {
		return namespace
	}
	return defaultNS
}
