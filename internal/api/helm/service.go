// Package helm Helm 관련 비즈니스 로직을 관리합니다.
package helm

import (
	"context"

	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
)

// Service : Helm 관련 비즈니스 로직 서비스
type Service struct{}

// NewService : 새로운 Helm 서비스 생성
func NewService() *Service {
	return &Service{}
}

// GetChartsInternal : Helm 차트 목록 조회 (내부 로직)
func (s *Service) GetChartsInternal(client client.Client, ctx context.Context, namespace string) (interface{}, error) {
	// Helm 차트 목록 조회
	charts, err := client.Helm().GetCharts(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return charts, nil
}

// GetChartInternal : 특정 Helm 차트 상세 조회 (내부 로직)
func (s *Service) GetChartInternal(client client.Client, ctx context.Context, chartName, namespace string, version int) (interface{}, error) {
	// 특정 차트 조회
	chart, err := client.Helm().GetChart(ctx, chartName, namespace, version)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

// GetChartStatusInternal : 차트 상태 조회 (내부 로직)
func (s *Service) GetChartStatusInternal(client client.Client, ctx context.Context, chartName, namespace string) (interface{}, error) {
	// 차트 상태 조회
	chart, err := client.Helm().GetChart(ctx, chartName, namespace, 0)
	if err != nil {
		return nil, err
	}

	// 상태 정보 구성
	status := map[string]interface{}{
		"name":        chart.Name,
		"namespace":   chart.Namespace,
		"status":      chart.Info.Status,
		"version":     chart.Version,
		"revision":    chart.Version,
		"chart":       chart.Chart.Metadata.Name,
		"app_version": chart.Chart.Metadata.AppVersion,
		"updated":     chart.Info.LastDeployed,
		"description": chart.Info.Description,
	}

	return status, nil
}

// InstallChartInternal : Helm 차트 설치 (내부 로직)
func (s *Service) InstallChartInternal(client client.Client, ctx context.Context, config config.InstallChartConfig) (interface{}, error) {
	// Helm 차트 설치
	err := client.Helm().InstallChart(config.ReleaseName, config.ChartURL, config.Version, config.Values)
	if err != nil {
		return nil, err
	}

	// 설치 결과 반환
	result := map[string]interface{}{
		"release_name": config.ReleaseName,
		"chart_url":    config.ChartURL,
		"version":      config.Version,
		"namespace":    config.Namespace,
		"status":       "installed",
	}

	return result, nil
}

// UpgradeChartInternal : Helm 차트 업그레이드 (내부 로직)
func (s *Service) UpgradeChartInternal(client client.Client, ctx context.Context, config config.UpgradeChartConfig) (interface{}, error) {
	// Helm 차트 업그레이드
	err := client.Helm().UpgradeChart(config.ReleaseName, config.ChartPath, config.Values)
	if err != nil {
		return nil, err
	}

	// 업그레이드 결과 반환
	result := map[string]interface{}{
		"release_name": config.ReleaseName,
		"chart_path":   config.ChartPath,
		"namespace":    config.Namespace,
		"status":       "upgraded",
	}

	return result, nil
}

// UninstallChartInternal : Helm 차트 제거 (내부 로직)
func (s *Service) UninstallChartInternal(client client.Client, ctx context.Context, releaseName, namespace string, dryRun bool) (interface{}, error) {
	// Helm 차트 제거
	err := client.Helm().UninstallChart(releaseName, namespace, dryRun)
	if err != nil {
		return nil, err
	}

	// 제거 결과 반환
	result := map[string]interface{}{
		"release_name": releaseName,
		"namespace":    namespace,
		"dry_run":      dryRun,
		"status":       "uninstalled",
	}

	return result, nil
}

// GetChartHistoryInternal : 차트 히스토리 조회 (내부 로직)
func (s *Service) GetChartHistoryInternal(client client.Client, ctx context.Context, chartName, namespace string) (interface{}, error) {
	// 차트 히스토리 조회 (최대 10개 버전)
	history := make([]interface{}, 0)
	for i := 1; i <= 10; i++ {
		chart, err := client.Helm().GetChart(ctx, chartName, namespace, i)
		if err != nil {
			// 더 이상 히스토리가 없으면 중단
			break
		}
		history = append(history, chart)
	}

	return map[string]interface{}{
		"chart_name": chartName,
		"namespace":  namespace,
		"history":    history,
		"count":      len(history),
	}, nil
}

// GetChartValuesInternal : 차트 값 조회 (내부 로직)
func (s *Service) GetChartValuesInternal(client client.Client, ctx context.Context, chartName, namespace string) (interface{}, error) {
	// 차트 조회 (최신 버전)
	chart, err := client.Helm().GetChart(ctx, chartName, namespace, 0)
	if err != nil {
		return nil, err
	}

	// Helm의 GetValues 액션을 사용하여 실제 values 조회
	values, err := client.Helm().GetValues(ctx, chartName, namespace)
	if err != nil {
		// GetValues 실패 시 차트의 Config 사용
		if chart.Config != nil {
			values = chart.Config
		} else {
			values = map[string]interface{}{}
		}
	}

	result := map[string]interface{}{
		"name":        chartName,
		"namespace":   namespace,
		"values":      values,
		"version":     chart.Version,
		"status":      chart.Info.Status,
		"chart_name":  chart.Chart.Metadata.Name,
		"app_version": chart.Chart.Metadata.AppVersion,
	}

	return result, nil
}

// HealthCheckInternal : Helm 연결 테스트 (내부 로직)
func (s *Service) HealthCheckInternal(client client.Client, ctx context.Context, namespace string) (interface{}, error) {
	// Helm 연결 테스트
	_, err := client.Helm().GetCharts(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"service":   "helm",
		"status":    "healthy",
		"namespace": namespace,
	}, nil
}
