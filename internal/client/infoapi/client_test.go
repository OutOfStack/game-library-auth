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
