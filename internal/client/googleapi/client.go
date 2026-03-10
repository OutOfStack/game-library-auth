package googleapi

import (
	"context"
	"net/http"
	"time"

	"github.com/OutOfStack/game-library-auth/pkg/observability"
	"go.opentelemetry.io/otel"
	"google.golang.org/api/idtoken"
)

var tracer = otel.Tracer("googleapi")

// Client represents Google OAuth API client
type Client struct {
	tokenValidator      *idtoken.Validator
	googleOAuthClientID string
}

// NewClient creates a new Google OAuth client
func NewClient(ctx context.Context, googleOAuthClientID string, timeout time.Duration) (*Client, error) {
	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: observability.NewTransport("google_api_client", observability.WithOtel()),
	}
	validator, err := idtoken.NewValidator(ctx, idtoken.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &Client{
		tokenValidator:      validator,
		googleOAuthClientID: googleOAuthClientID,
	}, nil
}

// ValidateIDToken validate google id token
func (c *Client) ValidateIDToken(ctx context.Context, idToken string) (*idtoken.Payload, error) {
	ctx, span := tracer.Start(ctx, "validateIDToken")
	defer span.End()

	return c.tokenValidator.Validate(ctx, idToken, c.googleOAuthClientID)
}
