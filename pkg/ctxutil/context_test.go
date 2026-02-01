package ctxutil_test

import (
	"context"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCtxWithTimeout_NoDeadline_AddsTimeout(t *testing.T) {
	timeout := 100 * time.Millisecond

	ctx, cancel := ctxutil.CtxWithTimeout(t.Context(), timeout)
	defer cancel()

	deadline, ok := ctx.Deadline()
	require.True(t, ok, "should have deadline set")
	assert.WithinDuration(t, time.Now().Add(timeout), deadline, 10*time.Millisecond)
}

func TestCtxWithTimeout_ExistingDeadline_PreservesDeadline(t *testing.T) {
	parentTimeout := 200 * time.Millisecond
	parentCtx, parentCancel := context.WithTimeout(t.Context(), parentTimeout)
	defer parentCancel()

	parentDeadline, _ := parentCtx.Deadline()

	ctx, cancel := ctxutil.CtxWithTimeout(parentCtx, 50*time.Millisecond)
	defer cancel()

	deadline, ok := ctx.Deadline()
	require.True(t, ok, "should have deadline")
	assert.Equal(t, parentDeadline, deadline, "should preserve parent deadline")
}

func TestCtxWithTimeout_ZeroTimeout_ReturnsCancel(t *testing.T) {
	ctx, cancel := ctxutil.CtxWithTimeout(t.Context(), 0)
	defer cancel()

	_, ok := ctx.Deadline()
	assert.False(t, ok, "should not have deadline with zero timeout")
}

func TestCtxWithTimeout_NegativeTimeout_ReturnsCancel(t *testing.T) {
	ctx, cancel := ctxutil.CtxWithTimeout(t.Context(), -1*time.Second)
	defer cancel()

	_, ok := ctx.Deadline()
	assert.False(t, ok, "should not have deadline with negative timeout")
}

func TestCtxWithTimeout_CancelFuncWorks(t *testing.T) {
	ctx, cancel := ctxutil.CtxWithTimeout(t.Context(), time.Hour)
	cancel()

	select {
	case <-ctx.Done():
		assert.Equal(t, context.Canceled, ctx.Err())
	default:
		t.Fatal("context should be canceled")
	}
}

func TestCtxWithTimeout_ExistingDeadline_CancelFuncWorks(t *testing.T) {
	parentCtx, parentCancel := context.WithTimeout(t.Context(), time.Hour)
	defer parentCancel()

	ctx, cancel := ctxutil.CtxWithTimeout(parentCtx, 50*time.Millisecond)
	cancel()

	select {
	case <-ctx.Done():
		assert.Equal(t, context.Canceled, ctx.Err())
	default:
		t.Fatal("context should be canceled")
	}
}
