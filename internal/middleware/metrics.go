package middleware

import (
	"strconv"
	"time"

	"github.com/OutOfStack/game-library-auth/pkg/observability"
	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpServerRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_server_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{observability.MethodLabel, observability.PathLabel, observability.CodeLabel},
	)
	httpServerRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_server_request_duration_seconds",
			Help:    "Histogram of response duration for HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{observability.MethodLabel, observability.PathLabel},
	)
	httpServerInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_server_in_flight_requests",
		Help: "Number of HTTP requests currently being processed",
	})
)

// Metrics records Prometheus metrics for each HTTP request
func Metrics() fiber.Handler {
	return func(c fiber.Ctx) error {
		// skip metrics endpoint to avoid self-instrumentation
		if c.Path() == "/metrics" {
			return c.Next()
		}

		httpServerInFlight.Inc()
		defer httpServerInFlight.Dec()

		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		path := c.Route().Path
		method := c.Method()
		code := strconv.Itoa(c.Response().StatusCode())

		httpServerRequestsTotal.WithLabelValues(method, path, code).Inc()
		httpServerRequestDuration.WithLabelValues(method, path).Observe(duration)

		return err
	}
}
