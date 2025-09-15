package utils

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/taking/kubemigrate/pkg/constants"
)

// TestNewWorkerPool : WorkerPool 생성자 테스트
// 새로 생성된 워커 풀이 올바르게 초기화되었는지 확인
func TestNewWorkerPool(t *testing.T) {
	workers := 3
	pool := NewWorkerPool(workers)

	if pool == nil {
		t.Fatal("NewWorkerPool() returned nil")
	}

	if pool.workers != workers {
		t.Errorf("NewWorkerPool() workers = %v, want %v", pool.workers, workers)
	}

	if pool.jobs == nil {
		t.Error("NewWorkerPool() jobs channel is nil")
	}

	if pool.results == nil {
		t.Error("NewWorkerPool() results channel is nil")
	}

	if pool.ctx == nil {
		t.Error("NewWorkerPool() context is nil")
	}

	if pool.cancel == nil {
		t.Error("NewWorkerPool() cancel function is nil")
	}
}

// TestWorkerPool_Submit - 작업 제출 테스트
// 워커 풀에 작업을 제출하고 실행되는지 확인
func TestWorkerPool_Submit(t *testing.T) {
	pool := NewWorkerPool(2)
	defer pool.Close()

	var executed bool
	var mu sync.Mutex

	// 작업 제출
	pool.Submit(func() {
		mu.Lock()
		executed = true
		mu.Unlock()
	})

	// 실행 대기
	time.Sleep(constants.TestWaitTimeLong)

	mu.Lock()
	if !executed {
		t.Error("Job was not executed")
	}
	mu.Unlock()
}

// TestWorkerPool_Submit_NonBlocking - 논블로킹 작업 제출 테스트
// 채널이 가득 찰 때 논블로킹으로 새 고루틴을 생성하는지 확인
func TestWorkerPool_Submit_NonBlocking(t *testing.T) {
	pool := NewWorkerPool(1)
	defer pool.Close()

	// 채널 가득 채움
	for i := 0; i < 3; i++ {
		pool.Submit(func() {
			time.Sleep(constants.TestWaitTimeLong)
		})
	}

	// 논블로킹 테스트
	start := time.Now()
	pool.Submit(func() {
		// 새 고루틴 실행
	})
	duration := time.Since(start)

	if duration > constants.TestWaitTimeShort {
		t.Error("Submit() should be non-blocking when channel is full")
	}
}

// TestWorkerPool_Close - 워커 풀 종료 테스트
// 워커 풀을 종료하고 컨텍스트가 취소되는지 확인
func TestWorkerPool_Close(t *testing.T) {
	pool := NewWorkerPool(2)

	// 작업 제출
	for i := 0; i < 5; i++ {
		pool.Submit(func() {
			time.Sleep(50 * time.Millisecond)
		})
	}

	// 풀 종료
	pool.Close()

	// 컨텍스트 취소 확인
	select {
	case <-pool.ctx.Done():
		// 정상
	default:
		t.Error("Context should be cancelled after Close()")
	}
}

// TestWorkerPool_Concurrent - 동시성 테스트
// 여러 작업이 동시에 실행되는지 확인
func TestWorkerPool_Concurrent(t *testing.T) {
	pool := NewWorkerPool(3)
	defer pool.Close()

	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 동시 작업 제출
	numJobs := 10
	wg.Add(numJobs)

	for i := 0; i < numJobs; i++ {
		pool.Submit(func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		})
	}

	wg.Wait()

	if counter != numJobs {
		t.Errorf("Expected %d jobs to execute, got %d", numJobs, counter)
	}
}

// TestWorkerPool_ContextCancellation - 컨텍스트 취소 테스트
// 컨텍스트가 취소된 후 작업이 실행되지 않는지 확인
func TestWorkerPool_ContextCancellation(t *testing.T) {
	pool := NewWorkerPool(2)

	// 컨텍스트 취소
	pool.cancel()

	// 워커 중지 대기
	time.Sleep(100 * time.Millisecond)

	// 취소 후 작업 제출
	var executed bool
	pool.Submit(func() {
		executed = true
	})

	// 대기
	time.Sleep(50 * time.Millisecond)

	// 실행되지 않아야 함
	if executed {
		t.Error("Job should not execute after context cancellation")
	}
}

// TestWithTimeout : 타임아웃 함수 테스트
// 타임아웃이 있는 함수 실행과 에러 처리 확인
func TestWithTimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		fn          func() error
		expectError bool
	}{
		{
			name:    "함수가 타임아웃 전에 완료",
			timeout: 100 * time.Millisecond,
			fn: func() error {
				time.Sleep(50 * time.Millisecond)
				return nil
			},
			expectError: false,
		},
		{
			name:    "함수가 타임아웃됨",
			timeout: 50 * time.Millisecond,
			fn: func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			},
			expectError: true,
		},
		{
			name:    "함수가 에러 반환",
			timeout: 100 * time.Millisecond,
			fn: func() error {
				return context.DeadlineExceeded
			},
			expectError: true,
		},
		{
			name:    "함수가 즉시 완료",
			timeout: 100 * time.Millisecond,
			fn: func() error {
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithTimeout(tt.timeout, tt.fn)
			if (err != nil) != tt.expectError {
				t.Errorf("WithTimeout() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestWorkerPool_JobChannel - 작업 채널 테스트
// 작업 채널의 버퍼 크기가 올바른지 확인
func TestWorkerPool_JobChannel(t *testing.T) {
	pool := NewWorkerPool(1)
	defer pool.Close()

	// 버퍼 크기 확인
	expectedBufferSize := pool.workers * 2
	actualBufferSize := cap(pool.jobs)

	if actualBufferSize != expectedBufferSize {
		t.Errorf("Jobs channel buffer size = %v, want %v", actualBufferSize, expectedBufferSize)
	}
}

// TestWorkerPool_ResultsChannel - 결과 채널 테스트
// 결과 채널의 버퍼 크기가 올바른지 확인
func TestWorkerPool_ResultsChannel(t *testing.T) {
	pool := NewWorkerPool(1)
	defer pool.Close()

	// 버퍼 크기 확인
	expectedBufferSize := pool.workers * 2
	actualBufferSize := cap(pool.results)

	if actualBufferSize != expectedBufferSize {
		t.Errorf("Results channel buffer size = %v, want %v", actualBufferSize, expectedBufferSize)
	}
}

// TestWorkerPool_WorkerGoroutines - 워커 고루틴 테스트
// 워커 고루틴들이 올바르게 작업을 처리하는지 확인
func TestWorkerPool_WorkerGoroutines(t *testing.T) {
	pool := NewWorkerPool(3)
	defer pool.Close()

	// 워커 시작 대기
	time.Sleep(constants.TestWaitTimeShort)

	// 작업 실행 확인
	var executedJobs int
	var mu sync.Mutex
	var wg sync.WaitGroup

	numJobs := 6
	wg.Add(numJobs)

	for i := 0; i < numJobs; i++ {
		pool.Submit(func() {
			defer wg.Done()
			mu.Lock()
			executedJobs++
			mu.Unlock()
		})
	}

	wg.Wait()

	if executedJobs != numJobs {
		t.Errorf("Expected %d jobs to execute, got %d", numJobs, executedJobs)
	}
}

// TestWorkerPool_CloseMultipleTimes - 여러 번 종료 테스트
// 워커 풀을 여러 번 종료해도 패닉이 발생하지 않는지 확인
func TestWorkerPool_CloseMultipleTimes(t *testing.T) {
	pool := NewWorkerPool(2)

	// 첫 번째 종료
	pool.Close()

	// 두 번째 종료 (무시됨)
	pool.Close()
}

// TestWorkerPool_SubmitAfterClose - 종료 후 작업 제출 테스트
// 워커 풀 종료 후 작업을 제출해도 안전하게 처리되는지 확인
func TestWorkerPool_SubmitAfterClose(t *testing.T) {
	pool := NewWorkerPool(1)
	pool.Close()

	// 종료 후 제출 (안전하게 처리)
	pool.Submit(func() {
		// 실행되지 않음
	})

	// 패닉 확인
	time.Sleep(constants.TestWaitTimeShort)
}
