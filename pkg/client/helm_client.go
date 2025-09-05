package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taking/kubemigrate/pkg/interfaces"
	"github.com/taking/kubemigrate/pkg/utils"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/taking/kubemigrate/pkg/models"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/client-go/rest"
)

// helmClient : Helm 클라이언트
type helmClient struct {
	cfg       *action.Configuration
	namespace string
	factory   *ClientFactory
}

// NewHelmClient : Helm 클라이언트 초기화
func NewHelmClient(cfg models.KubeConfig) (interfaces.HelmClient, error) {
	factory := NewClientFactory()

	restCfg, err := factory.CreateRESTConfig(cfg)
	if err != nil {
		return nil, err
	}

	ns := getNamespaceOrDefault(cfg.Namespace, "default")

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

	return utils.RunWithTimeout(ctx, func() error {
		_, err := list.Run()
		return err
	})
}

// GetCharts : Helm 차트 목록 조회
func (h *helmClient) GetCharts(ctx context.Context) ([]*release.Release, error) {
	list := action.NewList(h.cfg)
	list.AllNamespaces = true

	return list.Run()
}

// GetChart : Helm 차트 조회
func (h *helmClient) GetChart(ctx context.Context, releaseName string, namespace string, releaseVersion int) (*release.Release, error) {
	h.namespace = namespace

	get := action.NewGet(h.cfg)
	if releaseVersion != 0 {
		get.Version = releaseVersion
	}

	rel, err := get.Run(releaseName)
	if err != nil {
		return nil, err
	}
	return rel, nil
}

// IsChartInstalled : 특정 릴리스 설치 여부 확인 (모든 네임스페이스 검사)
func (h *helmClient) IsChartInstalled(releaseName string) (bool, *release.Release, error) {
	list := action.NewList(h.cfg)
	list.AllNamespaces = true

	releases, err := list.Run()
	if err != nil {
		return false, nil, fmt.Errorf("failed to list helm releases: %w", err)
	}

	for _, r := range releases {
		if r.Name == releaseName {
			return true, r, nil
		}
	}

	return false, nil, nil
}

// InstallChart : Helm 차트를 통해 설치, 설치 후 확인
func (h *helmClient) InstallChart(releaseName, chartPath string, values map[string]interface{}) error {

	// 설치 여부 확인 (릴리스 이름 기준)
	installed, _, err := h.IsChartInstalled(releaseName)
	if err != nil {
		return err
	}
	if installed {
		return fmt.Errorf("release '%s' is already installed", releaseName)
	}

	install := action.NewInstall(h.cfg)
	install.ReleaseName = releaseName
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
		return fmt.Errorf("failed to install release '%s': %w", releaseName, err)
	}

	// 설치 후 잠시 대기
	time.Sleep(2 * time.Second)

	// 설치 확인
	ok, _, err := h.IsChartInstalled(releaseName)
	if err != nil {
		return fmt.Errorf("failed to verify release '%s' installation: %w", releaseName, err)
	}
	if !ok {
		return fmt.Errorf("release '%s' installation not found after install", releaseName)
	}

	return nil
}

// UninstallChart : Helm 차트 삭제
func (h *helmClient) UninstallChart(releaseName, namespace string, dryRun bool) error {

	// 네임스페이스 설정
	h.namespace = namespace

	// 설치 여부 확인
	installed, _, err := h.IsChartInstalled(releaseName)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("chart '%s' is not installed (namespace: %s)", releaseName, namespace)
	}

	uninstall := action.NewUninstall(h.cfg)
	uninstall.DryRun = dryRun

	_, err = uninstall.Run(releaseName)
	if err != nil {
		if dryRun {
			return fmt.Errorf("[DryRun] failed to uninstall chart '%s' (namespace: %s): %w", releaseName, namespace, err)
		}
		return fmt.Errorf("failed to uninstall chart '%s' (namespace: %s): %w", releaseName, namespace, err)
	}

	// dryRun이 true인 경우 성공 메시지 반환
	if dryRun {
		return fmt.Errorf("[DryRun] chart '%s' uninstall simulation completed successfully (namespace: %s)", releaseName, namespace)
	}

	// dryRun=false 일 때만 삭제 확인
	if !dryRun {
		for i := 0; i < 5; i++ {
			installed, _, err := h.IsChartInstalled(releaseName)
			if err != nil {
				return fmt.Errorf("failed to verify chart '%s' uninstall: %w", releaseName, err)
			}
			if !installed {
				break
			}
			time.Sleep(2 * time.Second)
		}
	}

	// return nil

	// 3번 시도 후에도 여전히 설치되어 있으면 에러
	return fmt.Errorf("chart '%s' is still installed after uninstall (namespace: %s) - uninstall may be in progress", releaseName, namespace)
}
