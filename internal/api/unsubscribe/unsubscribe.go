package unsubscribe

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// UnsubscribeHandler handles GET /unsubscribe?token=xxx - shows confirmation page
func (a *API) UnsubscribeHandler(c fiber.Ctx) error {
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
