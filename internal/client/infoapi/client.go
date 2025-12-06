package infoapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	infoapipb "github.com/OutOfStack/game-library-auth/pkg/proto/infoapi/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultTimeout = 5 * time.Second

// Config - settings for infoapi service
type Config struct {
	Address     string
	Timeout     time.Duration
	DialOptions []grpc.DialOption
}

// Client wraps a gRPC InfoApiService client
type Client struct {
	cfg  Config
	conn *grpc.ClientConn
	api  infoapipb.InfoApiServiceClient
}

// NewClient dials the infoapi service and returns a ready client
func NewClient(_ context.Context, cfg Config) (*Client, error) {
	if cfg.Address == "" {
		return nil, errors.New("infoapi address is required")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}

	dialOpts := append([]grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, cfg.DialOptions...)

	conn, err := grpc.NewClient(cfg.Address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial infoapi: %w", err)
	}

	return &Client{
		cfg:  cfg,
		conn: conn,
		api:  infoapipb.NewInfoApiServiceClient(conn),
	}, nil
}

// Close closes the underlying gRPC connection
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}

	return c.conn.Close()
}

// CompanyExists checks if company exists in upstream infoapi service
func (c *Client) CompanyExists(ctx context.Context, companyName string) (bool, error) {
	if companyName == "" {
		return false, errors.New("company name is required")
	}

	ctx, cancel := withTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	req := &infoapipb.CompanyExistsRequest{}
	req.SetCompanyName(companyName)

	resp, err := c.api.CompanyExists(ctx, req)
	if err != nil {
		return false, fmt.Errorf("company exists: %w", err)
	}

	return resp.GetExists(), nil
}

func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok || timeout <= 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, timeout)
}
