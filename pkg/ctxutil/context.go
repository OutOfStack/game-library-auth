package ctxutil

import (
	"context"
	"time"
)

// CtxWithTimeout returns context with the provided timeout only if no deadline is already set,
// otherwise returns context with cancel function preserving the existing deadline
func CtxWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}

	if timeout <= 0 {
		return context.WithCancel(ctx)
	}

	return context.WithTimeout(ctx, timeout)
}
