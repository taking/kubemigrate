package helpers

import (
	"context"
	"time"
)

// TimeoutContext : 지정된 시간만큼 timeout context 생성
func TimeoutContext(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	if duration <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, duration)
}
