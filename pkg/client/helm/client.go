package helm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/validator"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client Helm 클라이언트 인터페이스
type Client interface {
	// Chart 관련
	GetCharts(ctx context.Context, namespace string) ([]*release.Release, error)
	GetChart(ctx context.Context, releaseName, namespace string, releaseVersion int) (*release.Release, error)
	IsChartInstalled(releaseName string) (bool, *release.Release, error)

	// Chart 설치/제거
	InstallChart(releaseName, chartPath string, values map[string]interface{}) error
	UninstallChart(releaseName, namespace string, dryRun bool) error
	UpgradeChart(releaseName, chartPath string, values map[string]interface{}) error

	// Health Check
	HealthCheck(ctx context.Context) error
}

// helmClient : Helm 클라이언트
type helmClient struct {
	cfg       *action.Configuration
	namespace string
}

// NewClient 새로운 Helm 클라이언트를 생성합니다 (기본 설정)
func NewClient() Client {
	// 기본 설정으로 클라이언트 생성
	actionConfig := new(action.Configuration)

	// 기본 네임스페이스 설정
	namespace := "default"

	// RESTClientGetter 생성
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		// in-cluster config가 없으면 기본 설정 사용
		restConfig = &rest.Config{}
	}

	flags := genericclioptions.NewConfigFlags(false)
	flags.Namespace = &namespace
	flags.WrapConfigFn = func(_ *rest.Config) *rest.Config {
		return restConfig
	}

	// Helm action.Configuration 초기화
	if err := actionConfig.Init(flags, namespace, "secret", log.Printf); err != nil {
		panic(fmt.Errorf("failed to initialize helm client: %w", err))
	}

	return &helmClient{
		cfg:       actionConfig,
		namespace: namespace,
	}
}

// NewClientWithConfig 설정을 받아서 Helm 클라이언트를 생성합니다
func NewClientWithConfig(cfg config.HelmConfig) (Client, error) {
	var restCfg *rest.Config
	var err error

	if cfg.KubeConfig.KubeConfig != "" {
		// 외부 kubeconfig 사용 (base64 디코딩)
		decodedKubeConfig, err := validator.DecodeIfBase64(cfg.KubeConfig.KubeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to decode kubeconfig: %w", err)
		}

		restCfg, err = clientcmd.RESTConfigFromKubeConfig([]byte(decodedKubeConfig))
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

	ns := getNamespaceOrDefault(cfg.KubeConfig.Namespace, "default")

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
	}, nil
}

// HealthCheck : Helm 연결 확인 (Kubernetes 연결 확인)
func (h *helmClient) HealthCheck(ctx context.Context) error {
	// 5초 제한 context 생성
	_, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	list := action.NewList(h.cfg)
	list.AllNamespaces = true

	_, err := list.Run()
	return err
}

// GetCharts : Helm 차트 목록 조회
func (h *helmClient) GetCharts(ctx context.Context, namespace string) ([]*release.Release, error) {
	list := action.NewList(h.cfg)
	list.AllNamespaces = true

	return list.Run()
}

// GetChart : Helm 차트 조회
func (h *helmClient) GetChart(ctx context.Context, releaseName, namespace string, releaseVersion int) (*release.Release, error) {
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

	// 3번 시도 후에도 여전히 설치되어 있으면 에러
	return fmt.Errorf("chart '%s' is still installed after uninstall (namespace: %s) - uninstall may be in progress", releaseName, namespace)
}

// UpgradeChart : Helm 차트 업그레이드
func (h *helmClient) UpgradeChart(releaseName, chartPath string, values map[string]interface{}) error {
	// 설치 여부 확인
	installed, _, err := h.IsChartInstalled(releaseName)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("release '%s' is not installed", releaseName)
	}

	upgrade := action.NewUpgrade(h.cfg)
	upgrade.Namespace = h.namespace

	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart: %w", err)
	}

	if values == nil {
		values = map[string]interface{}{}
	}

	_, err = upgrade.RunWithContext(context.Background(), releaseName, chart, values)
	if err != nil {
		return fmt.Errorf("failed to upgrade release '%s': %w", releaseName, err)
	}

	return nil
}

// getNamespaceOrDefault : 네임스페이스가 설정되어 있으면 사용, 없으면 기본값 반환
func getNamespaceOrDefault(namespace, defaultNS string) string {
	if namespace != "" {
		return namespace
	}
	return defaultNS
}
