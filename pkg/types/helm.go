package types

import (
	"fmt"

	"helm.sh/helm/v3/pkg/release"
)

// Helm 리소스 타입 정의
type (
	// Chart 관련
	ChartList     = []*release.Release
	Chart         = release.Release
	ChartResource interface {
		ChartList | *Chart
	}

	// Values 관련
	Values = map[string]interface{}
)

// Helm 타입 어설션 헬퍼 함수들

// 타입 어설션 헬퍼 함수들
func AssertChartList(v interface{}) (ChartList, bool) {
	chartList, ok := v.(ChartList)
	return chartList, ok
}

func AssertChart(v interface{}) (*Chart, bool) {
	chart, ok := v.(*Chart)
	return chart, ok
}

func AssertValues(v interface{}) (Values, bool) {
	values, ok := v.(Values)
	return values, ok
}

// 안전한 타입 어설션을 위한 래퍼 함수들
func SafeGetChartList(v interface{}) (ChartList, error) {
	if chartList, ok := AssertChartList(v); ok {
		return chartList, nil
	}
	return nil, fmt.Errorf("expected ChartList, got %T", v)
}

func SafeGetChart(v interface{}) (*Chart, error) {
	if chart, ok := AssertChart(v); ok {
		return chart, nil
	}
	return nil, fmt.Errorf("expected *Chart, got %T", v)
}

func SafeGetValues(v interface{}) (Values, error) {
	if values, ok := AssertValues(v); ok {
		return values, nil
	}
	return nil, fmt.Errorf("expected Values, got %T", v)
}
