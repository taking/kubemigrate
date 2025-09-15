// Package server 백그라운드 작업을 관리합니다.
package server

import (
	"time"

	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/logger"
	"github.com/taking/kubemigrate/pkg/constants"
	"github.com/taking/kubemigrate/pkg/utils"
)

// StartBackgroundTasks 백그라운드 작업들을 시작합니다.
func StartBackgroundTasks(baseHandler *handler.BaseHandler) {
	// 캐시 정리 작업
	go startCacheCleanup(baseHandler)

	// 메모리 모니터링 작업
	go startMemoryMonitoring()
}

// startCacheCleanup 캐시 정리 백그라운드 작업을 시작합니다.
func startCacheCleanup(baseHandler *handler.BaseHandler) {
	ticker := time.NewTicker(constants.CacheCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		baseHandler.CleanupCache()
	}
}

// startMemoryMonitoring 메모리 모니터링 백그라운드 작업을 시작합니다.
func startMemoryMonitoring() {
	utils.StartMemoryMonitor(5*time.Minute, 80.0, func(stats utils.MemoryStats) {
		// 메모리 사용량이 높을 때 로그 출력 및 최적화
		logger.Warn("High memory usage detected",
			logger.String("alloc", utils.FormatBytes(stats.Alloc)),
			logger.String("sys", utils.FormatBytes(stats.Sys)),
			logger.Any("usage_percent", utils.GetMemoryUsagePercent()),
			logger.Int("num_goroutines", stats.NumGoroutine),
		)
		utils.OptimizeMemory()
	})
}
