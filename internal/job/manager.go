package job

import (
	"fmt"
	"sync"
	"time"

	"github.com/taking/kubemigrate/pkg/utils"
)

// MemoryJobManager : 메모리 기반 작업 관리자 (워커 풀 통합)
type MemoryJobManager struct {
	jobs       map[string]*JobInfo
	mutex      sync.RWMutex
	workerPool *utils.WorkerPool
}

// NewMemoryJobManager : 메모리 작업 관리자 생성
func NewMemoryJobManager() *MemoryJobManager {
	return &MemoryJobManager{
		jobs:       make(map[string]*JobInfo),
		workerPool: utils.NewWorkerPool(5), // 기본 5개 워커
	}
}

// NewMemoryJobManagerWithWorkers : 워커 수를 지정하여 메모리 작업 관리자 생성
func NewMemoryJobManagerWithWorkers(workers int) *MemoryJobManager {
	return &MemoryJobManager{
		jobs:       make(map[string]*JobInfo),
		workerPool: utils.NewWorkerPool(workers),
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
		job.UpdatedAt = time.Now()
	}
}

// CompleteJob : 작업 완료
func (m *MemoryJobManager) CompleteJob(jobID string, result interface{}) {
	m.UpdateJobStatus(jobID, JobStatusCompleted, 100, "Job completed successfully")
	m.AddJobLog(jobID, "Job completed successfully")

	// 결과 저장
	m.mutex.Lock()
	if job, exists := m.jobs[jobID]; exists {
		job.Result = result
		job.UpdatedAt = time.Now()
	}
	m.mutex.Unlock()
}

// FailJob : 작업 실패
func (m *MemoryJobManager) FailJob(jobID string, err error) {
	m.UpdateJobStatus(jobID, JobStatusFailed, 0, err.Error())
	m.AddJobLog(jobID, fmt.Sprintf("Job failed: %s", err.Error()))
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

// DeleteJob : 작업 삭제
func (m *MemoryJobManager) DeleteJob(jobID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.jobs, jobID)
}

// RetryOperation : 재시도 로직이 포함된 작업 실행 (인터페이스 호환)
func (m *MemoryJobManager) RetryOperation(jobID, operationName string, maxAttempts int, operation func() error) error {
	return m.RetryOperationWithDelay(jobID, operationName, maxAttempts, 5*time.Second, operation)
}

// RetryOperationWithDelay : 재시도 로직이 포함된 작업 실행 (지연 포함)
func (m *MemoryJobManager) RetryOperationWithDelay(jobID, operationName string, maxAttempts int, delay time.Duration, operation func() error) error {
	return m.retryOperationInternal(jobID, operation, maxAttempts, delay, operationName)
}

// retryOperationInternal : 재시도 로직 내부 구현
func (m *MemoryJobManager) retryOperationInternal(jobID string, operation func() error, maxAttempts int, delay time.Duration, operationName string) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		m.UpdateJobStatus(jobID, JobStatusProcessing, (attempt-1)*100/maxAttempts, fmt.Sprintf("%s (attempt %d/%d)", operationName, attempt, maxAttempts))

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

// ExecuteJobAsync : 워커 풀을 사용하여 작업을 비동기로 실행
func (m *MemoryJobManager) ExecuteJobAsync(jobID string, task func() error) error {
	// 작업 상태를 실행 중으로 업데이트
	m.UpdateJobStatus(jobID, JobStatusProcessing, 0, "Task submitted to worker pool")

	// 워커 풀에 작업 제출
	m.workerPool.Submit(func() {
		// 작업 실행
		err := task()

		// 결과에 따라 상태 업데이트
		if err != nil {
			m.FailJob(jobID, err)
		} else {
			m.CompleteJob(jobID, "Task completed successfully")
		}
	})

	return nil
}

// ExecuteJobWithRetryAsync : 재시도 로직이 포함된 비동기 작업 실행
func (m *MemoryJobManager) ExecuteJobWithRetryAsync(jobID string, task func() error, maxAttempts int, delay time.Duration) error {
	// 작업 상태를 실행 중으로 업데이트
	m.UpdateJobStatus(jobID, JobStatusProcessing, 0, "Task submitted to worker pool with retry")

	// 워커 풀에 재시도 작업 제출
	m.workerPool.Submit(func() {
		// 재시도 로직 실행
		m.retryOperationInternal(jobID, task, maxAttempts, delay, "Async task")
	})

	return nil
}

// Close : 워커 풀 종료
func (m *MemoryJobManager) Close() error {
	if m.workerPool != nil {
		m.workerPool.Close()
	}
	return nil
}
