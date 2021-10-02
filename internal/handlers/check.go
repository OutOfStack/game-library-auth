package handlers

import (
	"net/http"

	"github.com/OutOfStack/game-library-auth/pkg/database"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// CheckAPI has methods for health checking
type CheckAPI struct {
	DB *sqlx.DB
}

type health struct {
	Status string `json:"status"`
}

// Health determines whether service is healthy
func (ca *CheckAPI) Health(c *fiber.Ctx) error {
	err := database.StatusCheck(ca.DB)
	h := &health{}
	if err != nil {
		h.Status = "database not ready"
		return c.Status(http.StatusInternalServerError).JSON(h)
	}
	h.Status = "OK"
	return c.JSON(h)
}
