package utils

import (
	"runtime"
	"runtime/debug"
	"time"
)

// MemoryStats : 메모리 사용량 통계
type MemoryStats struct {
	Alloc        uint64 `json:"alloc"`        // 현재 할당된 바이트 수
	TotalAlloc   uint64 `json:"totalAlloc"`   // 총 할당된 바이트 수
	Sys          uint64 `json:"sys"`          // 시스템에서 획득한 바이트 수
	NumGC        int32  `json:"numGC"`        // 가비지 컬렉션 횟수
	LastGC       int64  `json:"lastGC"`       // 마지막 GC 시간 (Unix nano)
	PauseTotalNs uint64 `json:"pauseTotalNs"` // 총 GC 일시정지 시간 (나노초)
	HeapInuse    uint64 `json:"heapInuse"`    // 힙 사용 중인 바이트 수
	HeapIdle     uint64 `json:"heapIdle"`     // 힙 유휴 바이트 수
	HeapReleased uint64 `json:"heapReleased"` // OS에 반환된 힙 바이트 수
	NumGoroutine int    `json:"numGoroutine"` // 현재 고루틴 수
}

// GetMemoryStats : 현재 메모리 사용량 통계 반환
func GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        int32(m.NumGC),
		LastGC:       int64(m.LastGC),
		PauseTotalNs: m.PauseTotalNs,
		HeapInuse:    m.HeapInuse,
		HeapIdle:     m.HeapIdle,
		HeapReleased: m.HeapReleased,
		NumGoroutine: runtime.NumGoroutine(),
	}
}

// ForceGC : 가비지 컬렉션 강제로 실행
func ForceGC() {
	runtime.GC()
}

// ForceGCAndFreeOS : 가비지 컬렉션 후 OS에 메모리 반환
func ForceGCAndFreeOS() {
	runtime.GC()
	debug.FreeOSMemory()
}

// GetMemoryUsagePercent : 메모리 사용률 백분율 반환
func GetMemoryUsagePercent() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 시스템 메모리 대비 사용률 계산
	if m.Sys > 0 {
		return float64(m.Alloc) / float64(m.Sys) * 100
	}
	return 0
}

// IsMemoryHigh : 메모리 사용량이 높은지 확인
func IsMemoryHigh(threshold float64) bool {
	return GetMemoryUsagePercent() > threshold
}

// StartMemoryMonitor : 메모리 모니터링 시작
func StartMemoryMonitor(interval time.Duration, highThreshold float64, callback func(MemoryStats)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		stats := GetMemoryStats()

		// 메모리 사용량이 높으면 콜백 호출
		if IsMemoryHigh(highThreshold) {
			callback(stats)
		}
	}
}

// OptimizeMemory : 메모리 최적화 수행
func OptimizeMemory() {
	// 가비지 컬렉션 강제 실행
	ForceGCAndFreeOS()

	// 메모리 사용량이 여전히 높으면 추가 최적화
	if IsMemoryHigh(80.0) {
		// 추가 최적화 로직 (예: 캐시 정리 등)
		ForceGC()
	}
}
