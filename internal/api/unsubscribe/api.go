package unsubscribe

import (
	"context"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("unsubscribeapi")

// API handles unsubscribe endpoints
type API struct {
	log               *zap.Logger
	tokenGenerator    *auth.UnsubscribeTokenGenerator
	unsubscribeFacade Facade
	contactEmail      string
}

// Facade provides methods for working with unsubscribe functionality
type Facade interface {
	UnsubscribeEmail(ctx context.Context, token string) (string, error)
	IsEmailUnsubscribed(ctx context.Context, email string) (bool, error)
}

// NewAPI creates a new unsubscribe API instance
func NewAPI(log *zap.Logger, tokenGenerator *auth.UnsubscribeTokenGenerator, unsubscribeFacade Facade, contactEmail string) *API {
	return &API{
		log:               log,
		tokenGenerator:    tokenGenerator,
		unsubscribeFacade: unsubscribeFacade,
		contactEmail:      contactEmail,
	}
}
