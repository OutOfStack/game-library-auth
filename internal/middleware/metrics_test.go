package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics_RecordsStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{name: "200 OK", statusCode: http.StatusOK},
		{name: "201 Created", statusCode: http.StatusCreated},
		{name: "400 Bad Request", statusCode: http.StatusBadRequest},
		{name: "404 Not Found", statusCode: http.StatusNotFound},
		{name: "500 Internal Server Error", statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(middleware.Metrics())
			app.Get("/test", func(c fiber.Ctx) error {
				return c.SendStatus(tt.statusCode)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.statusCode, resp.StatusCode)
		})
	}
}

func TestMetrics_DefaultStatusCode(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.Metrics())
	app.Get("/default", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/default", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMetrics_DifferentMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{name: "GET", method: http.MethodGet},
		{name: "POST", method: http.MethodPost},
		{name: "PATCH", method: http.MethodPatch},
		{name: "DELETE", method: http.MethodDelete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(middleware.Metrics())
			app.Add([]string{tt.method}, "/resource", func(c fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/resource", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestMetrics_SkipsMetricsEndpoint(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.Metrics())
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// /metrics endpoint should not be counted in http_server_requests_total
	assert.NotContains(t, string(body), `http_server_requests_total{code="200",method="GET",path="/metrics"}`)
}

func TestMetrics_ResponseBodyWritten(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.Metrics())
	app.Get("/body", func(c fiber.Ctx) error {
		return c.Status(http.StatusCreated).SendString("created")
	})

	req := httptest.NewRequest(http.MethodGet, "/body", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "created", string(body))
}

func TestMetrics_ExposedViaPrometheus(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.Metrics())
	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// make a request to generate metrics
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	// scrape metrics
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	metricsOutput := string(body)
	assert.Contains(t, metricsOutput, "http_server_requests_total")
	assert.Contains(t, metricsOutput, "http_server_request_duration_seconds")
	assert.Contains(t, metricsOutput, "http_server_in_flight_requests")
}
