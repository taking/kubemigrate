// Package constants 애플리케이션 전반에서 사용되는 상수들을 정의합니다.
package constants

import "time"

// 서버 관련 상수
const (
	// 기본 서버 설정
	DefaultServerHost = "localhost"
	DefaultServerPort = "9091"

	// 타임아웃 설정
	DefaultReadTimeout  = 30 * time.Second
	DefaultWriteTimeout = 30 * time.Second
	DefaultIdleTimeout  = 120 * time.Second

	// 요청 타임아웃
	DefaultRequestTimeout     = 30 * time.Second
	DefaultHealthCheckTimeout = 5 * time.Second

	// 서버 종료 타임아웃
	DefaultShutdownTimeout = 10 * time.Second
)

// 워커 풀 관련 상수
const (
	// 기본 워커 수
	DefaultWorkerPoolSize = 10

	// 채널 버퍼 크기 (워커 수의 2배)
	DefaultChannelBufferMultiplier = 2
)

// 캐시 관련 상수
const (
	// 기본 캐시 TTL
	DefaultCacheTTL = 5 * time.Minute

	// 테스트용 짧은 TTL
	TestCacheTTL = 50 * time.Millisecond

	// 캐시 정리 간격
	CacheCleanupInterval = 1 * time.Minute
)

// 메모리 관련 상수
const (
	// 메모리 모니터링 간격
	MemoryMonitoringInterval = 5 * time.Minute

	// 메모리 사용률 임계값
	MemoryUsageThreshold = 80.0
)

// 레이트 리미팅 관련 상수
const (
	// 기본 레이트 리미팅 설정
	DefaultRateLimitRate    = 100 // 초당 허용 요청 수
	DefaultRateLimitBurst   = 50  // 최대 버스트 수
	DefaultRateLimitExpires = 1 * time.Minute
)

// HTTP 클라이언트 관련 상수
const (
	// 기본 HTTP 타임아웃
	DefaultHTTPTimeout = 30 * time.Second
)

// Helm 관련 상수
const (
	// 차트 히스토리 최대 버전 수
	MaxChartHistoryVersions = 10
)

// 테스트 관련 상수
const (
	// 테스트 타임아웃
	TestTimeoutShort = 50 * time.Millisecond
	TestTimeoutLong  = 100 * time.Millisecond

	// 테스트 대기 시간
	TestWaitTimeShort = 50 * time.Millisecond
	TestWaitTimeLong  = 100 * time.Millisecond

	// 동시성 테스트용
	TestConcurrencyCount = 10
)

// KubeConfig 관련 상수
const (
	// KubeConfig 최대 크기 (100KB)
	MaxKubeConfigSize = 100000

	// 네임스페이스 최대 길이
	MaxNamespaceLength = 63
)
