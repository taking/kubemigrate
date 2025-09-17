package helm

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/taking/kubemigrate/internal/validator"
	"github.com/taking/kubemigrate/pkg/config"
	"github.com/taking/kubemigrate/pkg/constants"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client : Helm 클라이언트 인터페이스
type Client interface {
	// Chart 관련
	GetCharts(ctx context.Context, namespace string) ([]*release.Release, error)
	GetChart(ctx context.Context, releaseName, namespace string, releaseVersion int) (*release.Release, error)
	GetValues(ctx context.Context, releaseName, namespace string) (map[string]interface{}, error)
	IsChartInstalled(releaseName string) (bool, *release.Release, error)

	// Chart 설치/제거
	InstallChart(releaseName, chartURL, version string, values map[string]interface{}) error
	UninstallChart(releaseName, namespace string, dryRun bool) error
	UpgradeChart(releaseName, chartPath string, values map[string]interface{}) error
}

// helmClient : Helm 클라이언트
type helmClient struct {
	cfg       *action.Configuration
	namespace string
}

// NewClient : 새로운 Helm 클라이언트를 생성합니다 (기본 설정)
func NewClient() (Client, error) {
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
		return nil, fmt.Errorf("failed to initialize helm client: %w", err)
	}

	return &helmClient{
		cfg:       actionConfig,
		namespace: namespace,
	}, nil
}

// NewClientWithConfig : 설정을 받아서 Helm 클라이언트를 생성합니다
func NewClientWithConfig(cfg config.KubeConfig) (Client, error) {
	var restCfg *rest.Config
	var err error

	if cfg.KubeConfig != "" {
		// 외부 kubeconfig 사용 (base64 디코딩)
		decodedKubeConfig, err := validator.DecodeIfBase64(cfg.KubeConfig)
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

// GetValues : Helm 차트의 현재 values 조회
func (h *helmClient) GetValues(ctx context.Context, releaseName, namespace string) (map[string]interface{}, error) {
	h.namespace = namespace

	getValues := action.NewGetValues(h.cfg)
	getValues.AllValues = true // 모든 values 포함 (기본값 + 사용자 설정값)

	values, err := getValues.Run(releaseName)
	if err != nil {
		return nil, err
	}

	return values, nil
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

// InstallChart : Helm 차트를 URL에서 설치 (버전 지원)
func (h *helmClient) InstallChart(releaseName, chartURL, version string, values map[string]interface{}) error {
	// 설치 여부 확인 (릴리스 이름 기준)
	installed, _, err := h.IsChartInstalled(releaseName)
	if err != nil {
		return err
	}
	if installed {
		return fmt.Errorf("release '%s' is already installed", releaseName)
	}

	// chartURL이 URL인지 확인
	if !strings.HasPrefix(chartURL, "http://") && !strings.HasPrefix(chartURL, "https://") {
		return fmt.Errorf("chartURL must be a valid HTTP/HTTPS URL, got: %s", chartURL)
	}

	install := action.NewInstall(h.cfg)
	install.ReleaseName = releaseName
	install.Namespace = h.namespace
	install.Version = version

	// URL에서 차트 다운로드 및 로드
	chart, err := h.loadChartFromURL(chartURL, version)
	if err != nil {
		return fmt.Errorf("failed to load chart from URL: %w", err)
	}

	if values == nil {
		values = map[string]interface{}{}
	}

	_, err = install.Run(chart, values)
	if err != nil {
		return fmt.Errorf("failed to install release '%s': %w", releaseName, err)
	}

	// 설치 후 잠시 대기
	time.Sleep(constants.DefaultRequestTimeout)

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

// loadChartFromURL : URL에서 차트를 다운로드하고 로드
func (h *helmClient) loadChartFromURL(chartURL, version string) (*chart.Chart, error) {
	// 임시 디렉토리 생성
	tmpDir, err := os.MkdirTemp("", "helm-chart-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// 차트 다운로드
	chartPath, err := h.downloadChart(chartURL, version, tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to download chart: %w", err)
	}

	// 차트 로드
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load downloaded chart: %w", err)
	}

	return chart, nil
}

// downloadChart : 차트를 다운로드
func (h *helmClient) downloadChart(chartURL, version, destDir string) (string, error) {
	// HTTP 클라이언트 생성
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 차트 URL에 버전이 포함되어 있지 않으면 추가
	if version != "" && !strings.Contains(chartURL, version) {
		if strings.HasSuffix(chartURL, ".tgz") {
			// URL이 .tgz로 끝나는 경우 버전을 파일명에 추가
			baseURL := strings.TrimSuffix(chartURL, ".tgz")
			chartURL = fmt.Sprintf("%s-%s.tgz", baseURL, version)
		}
	}

	// 차트 다운로드
	resp, err := client.Get(chartURL)
	if err != nil {
		return "", fmt.Errorf("failed to download chart from %s: %w", chartURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download chart: HTTP %d", resp.StatusCode)
	}

	// 파일명 추출
	filename := filepath.Base(chartURL)
	if filename == "." || filename == "/" {
		filename = "chart.tgz"
	}

	// 파일 저장
	filePath := filepath.Join(destDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = file.Close() }()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save chart: %w", err)
	}

	return filePath, nil
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
			time.Sleep(constants.DefaultRequestTimeout)
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
