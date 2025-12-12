package unsubscribe

import (
	"errors"
	"net/http"

	"github.com/OutOfStack/game-library-auth/internal/database"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// UnsubscribeConfirmHandler handles POST /unsubscribe - processes the unsubscribe action
func (a *API) UnsubscribeConfirmHandler(c *fiber.Ctx) error {
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
