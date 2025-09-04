package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/taking/kubemigrate/pkg/models"
)

// HealthChecker 헬스체크를 수행하는 인터페이스
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}

// HealthManager 여러 헬스체커를 관리하는 구조체
type HealthManager struct {
	checkers map[string]HealthChecker
	cache    map[string]*models.HealthResponse
	mu       sync.RWMutex
	timeout  time.Duration
}

// NewHealthManager 새로운 헬스 매니저를 생성합니다
func NewHealthManager(timeout time.Duration) *HealthManager {
	return &HealthManager{
		checkers: make(map[string]HealthChecker),
		cache:    make(map[string]*models.HealthResponse),
		timeout:  timeout,
	}
}

// Register 헬스체커를 등록합니다
func (h *HealthManager) Register(checker HealthChecker) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.checkers[checker.Name()] = checker
}

// CheckAll : 모든 등록된 서비스의 헬스체크를 수행합니다
func (h *HealthManager) CheckAll(ctx context.Context) map[string]*models.HealthResponse {
	h.mu.RLock()
	checkers := make(map[string]HealthChecker)
	for name, checker := range h.checkers {
		checkers[name] = checker
	}
	h.mu.RUnlock()

	results := make(map[string]*models.HealthResponse)
	var wg sync.WaitGroup
	var resultMu sync.Mutex

	for name, checker := range checkers {
		wg.Add(1)
		go func(name string, checker HealthChecker) {
			defer wg.Done()

			result := h.checkSingle(ctx, name, checker)

			resultMu.Lock()
			results[name] = result
			resultMu.Unlock()
		}(name, checker)
	}

	wg.Wait()

	// 캐시 업데이트
	h.mu.Lock()
	for name, result := range results {
		h.cache[name] = result
	}
	h.mu.Unlock()

	return results
}

// CheckSingle : 특정 서비스의 헬스체크를 수행합니다
func (h *HealthManager) CheckSingle(ctx context.Context, name string) *models.HealthResponse {
	h.mu.RLock()
	checker, exists := h.checkers[name]
	h.mu.RUnlock()

	if !exists {
		return &models.HealthResponse{
			Status:    "unknown",
			Service:   name,
			Message:   "Service not registered",
			Timestamp: time.Now().UTC(),
		}
	}

	result := h.checkSingle(ctx, name, checker)

	// 캐시 업데이트
	h.mu.Lock()
	h.cache[name] = result
	h.mu.Unlock()

	return result
}

// checkSingle : 실제 헬스체크 수행 (내부 메서드)
func (h *HealthManager) checkSingle(ctx context.Context, name string, checker HealthChecker) *models.HealthResponse {
	// 타임아웃 컨텍스트 생성
	checkCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	start := time.Now()
	err := checker.Check(checkCtx)
	duration := time.Since(start)

	status := "healthy"
	message := "Service is healthy"

	if err != nil {
		status = "unhealthy"
		message = err.Error()
	}

	return &models.HealthResponse{
		Status:    status,
		Service:   name,
		Message:   message,
		Timestamp: time.Now().UTC(),
		Details: &struct {
			Version   string `json:"version,omitempty"`
			Namespace string `json:"namespace,omitempty"`
			Endpoint  string `json:"endpoint,omitempty"`
		}{
			// 체크 시간을 버전 필드에 임시 저장 (나중에 Duration 필드 추가 가능)
			Version: fmt.Sprintf("Check took %v", duration),
		},
	}
}

// GetCached 캐시된 헬스체크 결과를 반환합니다
func (h *HealthManager) GetCached(name string) (*models.HealthResponse, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result, exists := h.cache[name]
	return result, exists
}

// GetAllCached 모든 캐시된 헬스체크 결과를 반환합니다
func (h *HealthManager) GetAllCached() map[string]*models.HealthResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make(map[string]*models.HealthResponse)
	for name, result := range h.cache {
		results[name] = result
	}

	return results
}

// OverallHealth 전체 시스템의 헬스 상태를 반환합니다
func (h *HealthManager) OverallHealth(ctx context.Context) *models.HealthResponse {
	results := h.CheckAll(ctx)

	overallStatus := "healthy"
	var unhealthyServices []string

	for name, result := range results {
		if result.Status != "healthy" {
			overallStatus = "unhealthy"
			unhealthyServices = append(unhealthyServices, name)
		}
	}

	message := "All services are healthy"
	if overallStatus == "unhealthy" {
		message = fmt.Sprintf("Unhealthy services: %v", unhealthyServices)
	}

	return &models.HealthResponse{
		Status:    overallStatus,
		Service:   "system",
		Message:   message,
		Timestamp: time.Now().UTC(),
		Details: &struct {
			Version   string `json:"version,omitempty"`
			Namespace string `json:"namespace,omitempty"`
			Endpoint  string `json:"endpoint,omitempty"`
		}{
			Version: fmt.Sprintf("Total services: %d, Healthy: %d",
				len(results), len(results)-len(unhealthyServices)),
		},
	}
}
