package job

import (
	"time"
)

// JobStatus : 작업 상태
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// JobInfo : 작업 정보
type JobInfo struct {
	JobID     string                 `json:"jobId"`
	Status    JobStatus              `json:"status"`
	Progress  int                    `json:"progress"`
	Message   string                 `json:"message"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	Result    interface{}            `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Logs      []string               `json:"logs,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// JobManager : 작업 관리자 인터페이스
type JobManager interface {
	CreateJob(jobID string, metadata map[string]interface{}) *JobInfo
	UpdateJobStatus(jobID string, status JobStatus, progress int, message string)
	AddJobLog(jobID string, log string)
	CompleteJob(jobID string, result interface{})
	FailJob(jobID string, err error)
	GetJob(jobID string) (*JobInfo, bool)
	GetAllJobs() map[string]*JobInfo
	RetryOperation(jobID, operationName string, maxAttempts int, operation func() error) error
	RetryOperationWithDelay(jobID, operationName string, maxAttempts int, delay time.Duration, operation func() error) error
}
