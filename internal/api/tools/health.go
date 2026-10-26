package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/OutOfStack/game-library-auth/internal/version"
)

const (
	unavailable = "unavailable"
)

type health struct {
	version.Info
	Status    string `json:"status,omitempty"`
	Host      string `json:"host,omitempty"`
	Pod       string `json:"pod,omitempty"`
	PodIP     string `json:"podIP,omitempty"`
	Node      string `json:"node,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// Pinger checks database connectivity
type Pinger interface {
	PingContext(ctx context.Context) error
}

// HealthCheck has methods for readiness and liveness probes
type HealthCheck struct {
	db    Pinger
	ready atomic.Bool
}

// NewHealthCheck creates new HealthCheck
func NewHealthCheck(db Pinger) *HealthCheck {
	return &HealthCheck{db: db}
}

// SetReady marks whether service accepts traffic. Set it after listeners are bound and clear it on shutdown
func (hc *HealthCheck) SetReady(ready bool) {
	hc.ready.Store(ready)
}

// Readiness determines whether service is ready
func (hc *HealthCheck) Readiness(w http.ResponseWriter, r *http.Request) {
	h := newHealth()
	if !hc.ready.Load() {
		h.Status = "not ready"
		respond(w, h, http.StatusServiceUnavailable)
		return
	}
	if err := hc.db.PingContext(r.Context()); err != nil {
		h.Status = "database not ready"
		respond(w, h, http.StatusServiceUnavailable)
		return
	}
	h.Status = "OK"
	respond(w, h, http.StatusOK)
}

// Liveness determines whether service is up
func (hc *HealthCheck) Liveness(w http.ResponseWriter, _ *http.Request) {
	h := newHealth()
	h.Status = "OK"
	h.Pod = os.Getenv("KUBERNETES_PODNAME")
	h.PodIP = os.Getenv("KUBERNETES_PODIP")
	h.Node = os.Getenv("KUBERNETES_NODENAME")
	h.Namespace = os.Getenv("KUBERNETES_NAMESPACE")

	respond(w, h, http.StatusOK)
}

func newHealth() health {
	host, err := os.Hostname()
	if err != nil {
		host = unavailable
	}

	return health{
		Info: version.Get(),
		Host: host,
	}
}

func respond(w http.ResponseWriter, h health, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(h)
}
