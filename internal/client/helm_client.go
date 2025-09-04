package client

import (
	"context"
	"fmt"
	"github.com/taking/velero/internal/app"
	"github.com/taking/velero/internal/repository"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"log"
	"time"

	"github.com/taking/velero/internal/model"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/client-go/rest"
)

// HelmClient : Helm 클라이언트
type helmClient struct {
	cfg       *action.Configuration
	namespace string
	factory   *ClientFactory
}

// NewHelmClient : Helm 클라이언트 초기화
func NewHelmClient(cfg model.KubeConfig) (repository.HelmClient, error) {
	factory := NewClientFactory()

	restCfg, err := factory.CreateRESTConfig(cfg)
	if err != nil {
		return nil, err
	}

	ns := factory.ResolveNamespace(&cfg, "default")

	// genericclioptions.ConfigFlags 생성
	flags := genericclioptions.NewConfigFlags(false)
	flags.Namespace = &ns
	flags.WrapConfigFn = func(_ *rest.Config) *rest.Config {
		return restCfg
	}

	// Helm action.Configuration 초기화
	actionCfg := new(action.Configuration)
	if err := actionCfg.Init(flags, ns, "secret", log.Printf); err != nil {
		return nil, fmt.Errorf("failed to initialize helm client: %w", err)
	}

	return &helmClient{
		cfg:       actionCfg,
		namespace: ns,
		factory:   factory,
	}, nil
}

// HealthCheck : Helm 연결 확인 (Kubernetes 연결 확인)
func (h *helmClient) HealthCheck(ctx context.Context) error {
	// 5초 제한 context 생성
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	list := action.NewList(h.cfg)
	list.AllNamespaces = true

	return app.RunWithTimeout(ctx, func() error {
		_, err := list.Run()
		return err
	})
}

// IsChartInstalled : 특정 차트 설치 여부 확인 (모든 네임스페이스 검사)
func (h *helmClient) IsChartInstalled(chartName string) (bool, *release.Release, error) {
	list := action.NewList(h.cfg)
	list.AllNamespaces = true

	releases, err := list.Run()
	if err != nil {
		return false, nil, fmt.Errorf("failed to list helm releases: %w", err)
	}

	for _, r := range releases {
		if r.Chart != nil && r.Chart.Metadata.Name == chartName {
			return true, r, nil
		}
	}

	return false, nil, nil
}

// InstallChart : Helm 차트를 통해 설치, 설치 후 확인
func (h *helmClient) InstallChart(chartName, chartPath string, values map[string]interface{}) error {

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
		return fmt.Errorf("failed to load chart: %w", err)
	}

	if values == nil {
		values = map[string]interface{}{}
	}

	_, err = install.Run(chart, values)
	if err != nil {
		return fmt.Errorf("failed to install chart '%s': %w", chartName, err)
	}

	// 설치 후 잠시 대기
	time.Sleep(2 * time.Second)

	// 설치 확인
	ok, _, err := h.IsChartInstalled(chartName)
	if err != nil {
		return fmt.Errorf("failed to verify chart '%s' installation: %w", chartName, err)
	}
	if !ok {
		return fmt.Errorf("chart '%s' installation not found after install", chartName)
	}

	return nil
}

func (h *helmClient) InvalidateCache() error {
	dc, err := h.cfg.RESTClientGetter.ToDiscoveryClient()
	if err != nil {
		return fmt.Errorf("failed to get discovery client: %w", err)
	}

	if cachedDC, ok := dc.(interface{ Invalidate() }); ok {
		cachedDC.Invalidate()
		return nil
	}

	return fmt.Errorf("discovery client does not support Invalidate()")
}
