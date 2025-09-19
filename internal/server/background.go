// Package server 백그라운드 작업을 관리합니다.
package server

import (
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/logger"
)

// StartBackgroundTasks 백그라운드 작업들을 시작합니다.
func StartBackgroundTasks(baseHandler *handler.BaseHandler) {
	// 메모리 모니터링 작업
	go startMemoryMonitoring()
}

// startMemoryMonitoring 메모리 모니터링 백그라운드 작업을 시작합니다.
func startMemoryMonitoring() {
	// TODO: 메모리 모니터링 기능 구현 예정
	logger.Info("Memory monitoring not implemented yet")
}
