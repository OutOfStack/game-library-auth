package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/auth"
	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// UnsubscribeAPI handles unsubscribe endpoints
type UnsubscribeAPI struct {
	log               *zap.Logger
	tokenGenerator    *auth.UnsubscribeTokenGenerator
	unsubscribeFacade UnsubscribeFacade
	contactEmail      string
}

// UnsubscribeFacade provides methods for working with unsubscribe functionality
type UnsubscribeFacade interface {
	UnsubscribeEmail(ctx context.Context, token string) (string, error)
	IsEmailUnsubscribed(ctx context.Context, email string) (bool, error)
}

// NewUnsubscribeAPI creates a new unsubscribe API instance
func NewUnsubscribeAPI(log *zap.Logger, tokenGenerator *auth.UnsubscribeTokenGenerator, unsubscribeFacade UnsubscribeFacade, contactEmail string) *UnsubscribeAPI {
	return &UnsubscribeAPI{
		log:               log,
		tokenGenerator:    tokenGenerator,
		unsubscribeFacade: unsubscribeFacade,
		contactEmail:      contactEmail,
	}
}

// UnsubscribeHandler handles GET /unsubscribe?token=xxx - shows confirmation page
func (a *UnsubscribeAPI) UnsubscribeHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "unsubscribeHandler")
	defer span.End()

	token := c.Query("token")
	if token == "" {
		return c.Status(http.StatusBadRequest).SendString("Missing token parameter")
	}

	// validate token and extract email
	email, err := a.tokenGenerator.ValidateToken(token)
	if err != nil {
		a.log.Error("failed to validate unsubscribe token", zap.Error(err))
		return c.Status(http.StatusBadRequest).SendString("Invalid or expired unsubscribe link")
	}

	// check if email is already unsubscribed
	isUnsubscribed, err := a.unsubscribeFacade.IsEmailUnsubscribed(ctx, email)
	if err != nil {
		a.log.Error("failed to check if email is unsubscribed", zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString("Failed to check unsubscribe status")
	}

	if isUnsubscribed {
		// render already unsubscribed page
		return c.Render("unsubscribe_already", fiber.Map{
			"Email":        email,
			"ContactEmail": a.contactEmail,
		})
	}

	// render confirmation page
	return c.Render("unsubscribe", fiber.Map{
		"Email":        email,
		"Token":        token,
		"ContactEmail": a.contactEmail,
	})
}

// UnsubscribeConfirmHandler handles POST /unsubscribe - processes the unsubscribe action
func (a *UnsubscribeAPI) UnsubscribeConfirmHandler(c *fiber.Ctx) error {
	ctx, span := tracer.Start(c.Context(), "unsubscribeConfirmHandler")
	defer span.End()

	token := c.FormValue("token")
	if token == "" {
		return c.Status(http.StatusBadRequest).SendString("Missing token")
	}

	// process unsubscribe through facade
	email, err := a.unsubscribeFacade.UnsubscribeEmail(ctx, token)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return c.Status(http.StatusNotFound).SendString("Unsubscribe link not found or expired")
		}
		a.log.Error("failed to unsubscribe email", zap.Error(err))
		return c.Status(http.StatusInternalServerError).SendString("Failed to process unsubscribe request")
	}

	// render success page
	return c.Render("unsubscribe_success", fiber.Map{
		"Email":        email,
		"ContactEmail": a.contactEmail,
	})
}
