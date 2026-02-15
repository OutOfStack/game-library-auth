package observability

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// metrics labels
const (
	ClientLabel = "client"
	URLLabel    = "url"
	MethodLabel = "method"
	CodeLabel   = "code"
)

const (
	maxURLSegments = 4
)

var (
	httpClientInFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_client_in_flight_requests",
		Help: "A gauge of in-flight requests for the HTTP client",
	}, []string{ClientLabel})

	httpClientRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_client_requests_total",
		Help: "Total number of HTTP client requests",
	}, []string{ClientLabel, MethodLabel, CodeLabel, URLLabel})

	httpClientRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_client_request_duration_seconds",
		Help:    "A histogram of HTTP client request latencies",
		Buckets: prometheus.DefBuckets,
	}, []string{ClientLabel, MethodLabel, URLLabel})

	httpClientRequestErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_client_request_errors_total",
		Help: "Total number of HTTP client transport errors",
	}, []string{ClientLabel, MethodLabel, URLLabel})
)

// TransportOption configures a monitoredTransport
type TransportOption func(*transportOptions)

type transportOptions struct {
	rt       http.RoundTripper
	withOtel bool
}

// WithRoundTripper sets the base RoundTripper to wrap
func WithRoundTripper(rt http.RoundTripper) TransportOption {
	return func(o *transportOptions) {
		o.rt = rt
	}
}

// WithOtel wraps the transport with OpenTelemetry instrumentation for trace propagation
func WithOtel() TransportOption {
	return func(o *transportOptions) {
		o.withOtel = true
	}
}

// transport wraps an http.RoundTripper to add metrics
type transport struct {
	rt         http.RoundTripper
	clientName string
}

// NewTransport creates a new instrumented http.RoundTripper.
// It is safe to call multiple times with the same namespace/subsystem - subsequent
// calls will return the same transport instance to avoid duplicate metric registration.
func NewTransport(clientName string, opts ...TransportOption) http.RoundTripper {
	options := &transportOptions{
		rt: http.DefaultTransport,
	}
	for _, opt := range opts {
		opt(options)
	}

	rt := options.rt
	if options.withOtel {
		rt = otelhttp.NewTransport(rt)
	}

	return &transport{
		rt:         rt,
		clientName: clientName,
	}
}

// RoundTrip implements the http.RoundTripper interface
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	labelURL := normalizeURL(*req.URL, maxURLSegments)
	method := req.Method

	httpClientInFlight.WithLabelValues(t.clientName).Inc()
	defer httpClientInFlight.WithLabelValues(t.clientName).Dec()

	start := time.Now()
	resp, err := t.rt.RoundTrip(req)
	duration := time.Since(start).Seconds()

	httpClientRequestDuration.WithLabelValues(t.clientName, method, labelURL).Observe(duration)

	if err != nil {
		httpClientRequestErrors.WithLabelValues(t.clientName, method, labelURL).Inc()
	} else if resp != nil {
		httpClientRequestsTotal.WithLabelValues(t.clientName, method, strconv.Itoa(resp.StatusCode), labelURL).Inc()
	}

	return resp, err
}

func normalizeURL(u url.URL, maxSegments int) string {
	parts := strings.Split(u.Path, "/")
	var segments []string
	for _, part := range parts {
		if part != "" {
			segments = append(segments, part)
		}
	}

	if len(segments) > maxSegments {
		segments = segments[:maxSegments]
	}

	u.Path = "/" + strings.Join(segments, "/")
	u.RawQuery = ""

	return u.String()
}
