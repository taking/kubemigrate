package utils

import (
	"context"
	"fmt"
)

// RunWithTimeout : 컨텍스트 타임아웃과 함께 함수 실행
func RunWithTimeout(ctx context.Context, fn func() error) error {
	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("operation failed: timeout")
	case err := <-done:
		if err != nil {
			return fmt.Errorf("operation failed: %w", err)
		}
		return nil
	}
}
