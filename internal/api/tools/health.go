package tools

import (
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

const (
	unavailable = "unavailable"
)

type health struct {
	Status    string `json:"status,omitempty"`
	Host      string `json:"host,omitempty"`
	Pod       string `json:"pod,omitempty"`
	PodIP     string `json:"podIP,omitempty"`
	Node      string `json:"node,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// HealthCheckAPI has methods for readiness and liveness checking
type HealthCheckAPI struct {
	db *sqlx.DB
}

// NewHealthCheckAPI returns new instance of healthcheck api
func NewHealthCheckAPI(db *sqlx.DB) *HealthCheckAPI {
	return &HealthCheckAPI{
		db: db,
	}
}

// Readiness determines whether service is ready
func (a *HealthCheckAPI) Readiness(c *fiber.Ctx) error {
	var h health
	host, err := os.Hostname()
	if err != nil {
		host = unavailable
	}
	h.Host = host
	err = a.db.PingContext(c.Context())
	if err != nil {
		h.Status = "database not ready"
		return c.Status(http.StatusInternalServerError).JSON(h)
	}
	h.Status = "OK"
	return c.JSON(h)
}

// Liveness determines whether service is up
func (a *HealthCheckAPI) Liveness(c *fiber.Ctx) error {
	host, err := os.Hostname()
	if err != nil {
		host = unavailable
	}
	h := health{
		Host:      host,
		Status:    "OK",
		Pod:       os.Getenv("KUBERNETES_PODNAME"),
		PodIP:     os.Getenv("KUBERNETES_PODIP"),
		Node:      os.Getenv("KUBERNETES_NODENAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}

	return c.JSON(h)
}
