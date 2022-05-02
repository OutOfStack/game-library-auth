package handlers

import (
	"net/http"
	"os"

	"github.com/OutOfStack/game-library-auth/pkg/database"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

// CheckAPI has methods for readiness and liveness checking
type CheckAPI struct {
	DB *sqlx.DB
}

type health struct {
	Status string `json:"status"`
	Host   string `json:"host"`
}

// Readiness determines whether service is ready
func (ca *CheckAPI) Readiness(c *fiber.Ctx) error {
	var h health
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}
	h.Host = host
	err = database.StatusCheck(ca.DB)
	if err != nil {
		h.Status = "database not ready"
		return c.Status(http.StatusInternalServerError).JSON(h)
	}
	h.Status = "OK"
	return c.JSON(h)
}

// Liveness determines whether service is up
func (ca *CheckAPI) Liveness(c *fiber.Ctx) error {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}
	h := health{
		Host:   host,
		Status: "OK",
	}

	return c.JSON(h)
}
