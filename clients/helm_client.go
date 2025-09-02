package clients

import (
	"context"
	"fmt"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/restmapper"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"log"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"taking.kr/velero/models"
)

// HelmClient : Helm 클라이언트
type HelmClient struct {
	cfg       *action.Configuration
	namespace string
}

// NewHelmClient : Helm 클라이언트 초기화
func NewHelmClient(cfg models.KubeConfig) (*HelmClient, error) {
	var restCfg *rest.Config
	var err error

	// 없으면 in-cluster config 시도
	if cfg.KubeConfig != "" {
		restCfg, err = clientcmd.RESTConfigFromKubeConfig([]byte(cfg.KubeConfig))
		if err != nil {
			return nil, fmt.Errorf("❌ failed to parse kubeconfig: %w", err)
		}
	} else {
		// 없으면 in-cluster config 시도
		restCfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("❌ failed to load in-cluster config: %w", err)
		}
	}

	// 네임스페이스 없으면 기본 "default"
	ns := cfg.Namespace
	if ns == "" {
		ns = "default"
	}

	// genericclioptions.ConfigFlags 생성
	flags := genericclioptions.NewConfigFlags(false)
	flags.Namespace = &ns
	flags.WrapConfigFn = func(_ *rest.Config) *rest.Config {
		return restCfg
	}

	// Helm action.Configuration 초기화
	actionCfg := new(action.Configuration)
	if err := actionCfg.Init(flags, ns, "secret", log.Printf); err != nil {
		return nil, fmt.Errorf("❌ failed to initialize helm client: %w", err)
	}

	return &HelmClient{
		cfg:       actionCfg,
		namespace: ns,
	}, nil
}

// HealthCheck : Helm 연결 확인 (Kubernetes 연결 확인)
func (h *HelmClient) HealthCheck() error {
	// 5초 제한 context 생성
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	list := action.NewList(h.cfg)
	list.AllNamespaces = true

	done := make(chan error, 1)
	go func() {
		_, err := list.Run()
		done <- err
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("❌ failed to helm health check: timeout")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("❌ failed to helm health check: %w", err)
		}
	}

	return nil
}

// IsChartInstalled : 특정 차트 설치 여부 확인 (모든 네임스페이스 검사)
func (h *HelmClient) IsChartInstalled(chartName string) (bool, *release.Release, error) {
	list := action.NewList(h.cfg)
	list.AllNamespaces = true

	releases, err := list.Run()
	if err != nil {
		return false, nil, fmt.Errorf("❌ failed to list helm releases: %w", err)
	}

	for _, r := range releases {
		if r.Chart != nil && r.Chart.Metadata.Name == chartName {
			return true, r, nil
		}
	}

	return false, nil, nil
}

// InstallChart : Helm 차트를 통해 설치, 설치 후 확인
func (h *HelmClient) InstallChart(chartName, chartPath string, values map[string]interface{}) error {

	// 설치 여부 확인 (릴리스 이름이 고정이 아니므로 chart 기준)
	installed, _, err := h.IsChartInstalled(chartName)
	if err != nil {
		return err
	}
	if installed {
		return fmt.Errorf("chart '%s' is already installed", chartName)
	}

	install := action.NewInstall(h.cfg)
	install.ReleaseName = chartName
	install.Namespace = h.namespace

	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("❌ failed to load chart: %w", err)
	}

	if values == nil {
		values = map[string]interface{}{}
	}

	_, err = install.Run(chart, values)
	if err != nil {
		return fmt.Errorf("❌ failed to install chart '%s': %w", chartName, err)
	}

	// 설치 후 잠시 대기
	time.Sleep(2 * time.Second)

	// 설치 확인
	ok, _, err := h.IsChartInstalled(chartName)
	if err != nil {
		return fmt.Errorf("❌ failed to verify chart '%s' installation: %w", chartName, err)
	}
	if !ok {
		return fmt.Errorf("❌ chart '%s' installation not found after install", chartName)
	}

	return nil
}

// restGetter : Helm action.Configuration Init 용 RESTClientGetter
type restGetter struct {
	restCfg *rest.Config
}

func (r restGetter) ToRESTConfig() (*rest.Config, error) {
	return r.restCfg, nil
}

// ToDiscoveryClient : discovery.CachedDiscoveryInterface 반환
func (r restGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(r.restCfg)
	if err != nil {
		return nil, err
	}
	return memory.NewMemCacheClient(dc), nil
}

// ToRESTMapper : RESTMapper 반환
func (r restGetter) ToRESTMapper() (meta.RESTMapper, error) {
	dc, err := r.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	return restmapper.NewDeferredDiscoveryRESTMapper(dc), nil
}

// ToRawKubeConfigLoader : kubeconfig loader 반환
func (r restGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return clientcmd.NewDefaultClientConfig(clientcmdapi.Config{}, &clientcmd.ConfigOverrides{})
}

// InvalidateCache : Helm client discovery cache 초기화 (항상 최신 상태 반영)
func (h *HelmClient) InvalidateCache() error {
	dc, err := h.cfg.RESTClientGetter.ToDiscoveryClient()
	if err != nil {
		return fmt.Errorf("failed to get discovery client: %w", err)
	}

	// 캐시 초기화
	if cachedDC, ok := dc.(interface{ Invalidate() }); ok {
		cachedDC.Invalidate()
		return nil
	}

	return fmt.Errorf("discovery client does not support Invalidate()")
}
