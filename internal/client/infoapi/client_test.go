package infoapi_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/client/infoapi"
	infoapipb "github.com/OutOfStack/game-library-auth/pkg/proto/infoapi/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type infoapiServer struct {
	infoapipb.UnimplementedInfoApiServiceServer
	exists bool
}

func (s *infoapiServer) CompanyExists(_ context.Context, _ *infoapipb.CompanyExistsRequest) (*infoapipb.CompanyExistsResponse, error) {
	resp := &infoapipb.CompanyExistsResponse{}
	resp.SetExists(s.exists)
	return resp, nil
}

func TestNewClientRequiresAddress(t *testing.T) {
	_, err := infoapi.NewClient(t.Context(), infoapi.Config{})
	require.Error(t, err)
}

func TestCompanyExists(t *testing.T) {
	addr, cleanup := startServer(t, true)
	defer cleanup()

	client, err := infoapi.NewClient(t.Context(), infoapi.Config{
		Address: addr,
		Timeout: time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, client.Close())
	})

	exists, err := client.CompanyExists(t.Context(), "Nintendo")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestCompanyExistsRequiresName(t *testing.T) {
	addr, cleanup := startServer(t, true)
	defer cleanup()

	client, err := infoapi.NewClient(t.Context(), infoapi.Config{
		Address: addr,
		Timeout: time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, client.Close())
	})

	_, err = client.CompanyExists(t.Context(), "")
	require.Error(t, err)
}

func TestCtxWithTimeout(t *testing.T) {
	t.Run("sets timeout when context has no deadline", func(t *testing.T) {
		ctx := context.Background()
		timeout := 100 * time.Millisecond

		newCtx, cancel := infoapi.CtxWithTimeout(ctx, timeout)
		defer cancel()

		deadline, ok := newCtx.Deadline()
		require.True(t, ok)
		assert.WithinDuration(t, time.Now().Add(timeout), deadline, 50*time.Millisecond)
	})

	t.Run("respects existing deadline", func(t *testing.T) {
		existingDeadline := time.Now().Add(200 * time.Millisecond)
		ctx, cancel := context.WithDeadline(context.Background(), existingDeadline)
		defer cancel()

		newCtx, newCancel := infoapi.CtxWithTimeout(ctx, 100*time.Millisecond)
		defer newCancel()

		deadline, ok := newCtx.Deadline()
		require.True(t, ok)
		assert.Equal(t, existingDeadline, deadline)
	})

	t.Run("returns working cancel function when deadline exists", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		newCtx, newCancel := infoapi.CtxWithTimeout(ctx, 100*time.Millisecond)
		require.NotNil(t, newCancel)

		newCancel()

		select {
		case <-newCtx.Done():
		case <-time.After(100 * time.Millisecond):
			t.Fatal("context should be cancelled")
		}
	})

	t.Run("handles zero timeout", func(t *testing.T) {
		ctx := context.Background()

		newCtx, cancel := infoapi.CtxWithTimeout(ctx, 0)
		defer cancel()

		_, ok := newCtx.Deadline()
		assert.False(t, ok)
	})

	t.Run("handles negative timeout", func(t *testing.T) {
		ctx := context.Background()

		newCtx, cancel := infoapi.CtxWithTimeout(ctx, -1*time.Second)
		defer cancel()

		_, ok := newCtx.Deadline()
		assert.False(t, ok)
	})

	t.Run("cancel function can be called multiple times", func(_ *testing.T) {
		ctx := context.Background()

		_, cancel := infoapi.CtxWithTimeout(ctx, 100*time.Millisecond)

		cancel()
		cancel()
	})
}

func startServer(t *testing.T, exists bool) (string, func()) {
	t.Helper()

	lc := &net.ListenConfig{}
	lis, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	infoapipb.RegisterInfoApiServiceServer(srv, &infoapiServer{exists: exists})

	go func() {
		_ = srv.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		srv.GracefulStop()
		_ = lis.Close()
	}
}
