package utils

import (
	"context"
	"sync"
	"time"
)

// WorkerPool : 고루틴 풀을 관리하는 구조체
type WorkerPool struct {
	workers int
	jobs    chan func()
	results chan interface{}
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewWorkerPool : 새로운 워커 풀을 생성합니다
func NewWorkerPool(workers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers: workers,
		jobs:    make(chan func(), workers*2), // 버퍼 크기를 워커 수의 2배로 설정
		results: make(chan interface{}, workers*2),
		ctx:     ctx,
		cancel:  cancel,
	}

	// 워커들 시작
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

// worker 개별 워커 고루틴
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case job := <-p.jobs:
			if job != nil {
				job()
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// Submit 작업을 워커 풀에 제출합니다
func (p *WorkerPool) Submit(job func()) {
	// 컨텍스트가 이미 취소된 경우 무시
	select {
	case <-p.ctx.Done():
		return
	default:
	}

	select {
	case p.jobs <- job:
	case <-p.ctx.Done():
		return
	default:
		// 논블로킹: 풀이 가득 차면 새 고루틴으로 실행
		go job()
	}
}

// Close 워커 풀을 종료합니다
func (p *WorkerPool) Close() {
	p.cancel()

	// jobs 채널이 이미 닫혀있지 않은 경우에만 닫기
	select {
	case <-p.jobs:
		// 이미 닫혀있음
	default:
		close(p.jobs)
	}

	p.wg.Wait()

	// results 채널이 이미 닫혀있지 않은 경우에만 닫기
	select {
	case <-p.results:
		// 이미 닫혀있음
	default:
		close(p.results)
	}
}

// WithTimeout 타임아웃과 함께 작업을 실행합니다
func WithTimeout(timeout time.Duration, fn func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
