// Package config 애플리케이션 설정을 관리합니다.
package config

import (
	pkgconfig "github.com/taking/kubemigrate/pkg/config"
)

// Load : 환경변수 기반으로 Config 구조체 생성
// 환경변수가 없으면 기본값 사용
func Load() *pkgconfig.Config {
	return pkgconfig.Load()
}
