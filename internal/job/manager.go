package job

import (
	"fmt"
	"sync"
	"time"
)

// MemoryJobManager : 메모리 기반 작업 관리자
type MemoryJobManager struct {
	jobs  map[string]*JobInfo
	mutex sync.RWMutex
}

// NewMemoryJobManager : 메모리 작업 관리자 생성
func NewMemoryJobManager() *MemoryJobManager {
	return &MemoryJobManager{
		jobs: make(map[string]*JobInfo),
	}
}

// CreateJob : 작업 생성
func (m *MemoryJobManager) CreateJob(jobID string, metadata map[string]interface{}) *JobInfo {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	job := &JobInfo{
		JobID:     jobID,
		Status:    JobStatusPending,
		Progress:  0,
		Message:   "Job created",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  metadata,
		Logs:      []string{},
	}

	m.jobs[jobID] = job
	return job
}

// UpdateJobStatus : 작업 상태 업데이트
func (m *MemoryJobManager) UpdateJobStatus(jobID string, status JobStatus, progress int, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if job, exists := m.jobs[jobID]; exists {
		job.Status = status
		job.Progress = progress
		job.Message = message
		job.UpdatedAt = time.Now()
	}
}

// AddJobLog : 작업 로그 추가
func (m *MemoryJobManager) AddJobLog(jobID string, log string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if job, exists := m.jobs[jobID]; exists {
		job.Logs = append(job.Logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), log))
	}
}

// CompleteJob : 작업 완료
func (m *MemoryJobManager) CompleteJob(jobID string, result interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if job, exists := m.jobs[jobID]; exists {
		job.Status = JobStatusCompleted
		job.Progress = 100
		job.Message = "Job completed successfully"
		job.Result = result
		job.UpdatedAt = time.Now()
	}
}

// FailJob : 작업 실패
func (m *MemoryJobManager) FailJob(jobID string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if job, exists := m.jobs[jobID]; exists {
		job.Status = JobStatusFailed
		job.Message = "Job failed"
		job.Error = err.Error()
		job.UpdatedAt = time.Now()
	}
}

// GetJob : 작업 조회
func (m *MemoryJobManager) GetJob(jobID string) (*JobInfo, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	job, exists := m.jobs[jobID]
	return job, exists
}

// GetAllJobs : 모든 작업 조회
func (m *MemoryJobManager) GetAllJobs() map[string]*JobInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 복사본 반환
	jobs := make(map[string]*JobInfo)
	for k, v := range m.jobs {
		jobs[k] = v
	}
	return jobs
}

// RetryOperation : 재시도 로직을 포함한 작업 실행
func (m *MemoryJobManager) RetryOperation(jobID, operationName string, maxAttempts int, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		m.AddJobLog(jobID, fmt.Sprintf("%s attempt %d/%d", operationName, attempt, maxAttempts))

		err := operation()
		if err == nil {
			m.AddJobLog(jobID, fmt.Sprintf("%s completed successfully", operationName))
			return nil
		}

		lastErr = err
		if attempt < maxAttempts {
			m.AddJobLog(jobID, fmt.Sprintf("%s failed (attempt %d): %s. Retrying in 5 seconds...", operationName, attempt, err.Error()))
			time.Sleep(5 * time.Second)
		} else {
			// 마지막 시도에서 실패한 경우 로그 추가
			m.AddJobLog(jobID, fmt.Sprintf("%s failed (attempt %d): %s", operationName, attempt, err.Error()))
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", operationName, maxAttempts, lastErr)
}

// RetryOperationWithDelay : 사용자 정의 지연 시간을 가진 재시도 로직
func (m *MemoryJobManager) RetryOperationWithDelay(jobID, operationName string, maxAttempts int, delay time.Duration, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		m.AddJobLog(jobID, fmt.Sprintf("%s attempt %d/%d", operationName, attempt, maxAttempts))

		err := operation()
		if err == nil {
			m.AddJobLog(jobID, fmt.Sprintf("%s completed successfully", operationName))
			return nil
		}

		lastErr = err
		if attempt < maxAttempts {
			m.AddJobLog(jobID, fmt.Sprintf("%s failed (attempt %d): %s. Retrying in %v...", operationName, attempt, err.Error(), delay))
			time.Sleep(delay)
		} else {
			// 마지막 시도에서 실패한 경우 로그 추가
			m.AddJobLog(jobID, fmt.Sprintf("%s failed (attempt %d): %s", operationName, attempt, err.Error()))
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", operationName, maxAttempts, lastErr)
}
