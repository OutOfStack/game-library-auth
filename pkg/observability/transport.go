package observability

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	transportsMu sync.Mutex
	transports   = make(map[string]http.RoundTripper)
)

// NewTransport creates a new instrumented http.RoundTripper.
// It is safe to call multiple times with the same namespace/subsystem - subsequent
// calls will return the same transport instance to avoid duplicate metric registration.
func NewTransport(namespace, subsystem string) http.RoundTripper {
	key := namespace + "/" + subsystem

	transportsMu.Lock()
	defer transportsMu.Unlock()
	if t, ok := transports[key]; ok {
		return t
	}

	inFlight := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "client_in_flight_requests",
		Help:      "A gauge of in-flight requests for the client.",
	})

	counter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "client_api_requests_total",
			Help:      "A counter for requests from the client.",
		},
		[]string{"code", "method"},
	)

	histVec := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "client_api_request_duration_seconds",
			Help:      "A histogram of request latencies.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	promTransport := promhttp.InstrumentRoundTripperInFlight(inFlight,
		promhttp.InstrumentRoundTripperCounter(counter,
			promhttp.InstrumentRoundTripperDuration(histVec, http.DefaultTransport),
		),
	)

	t := otelhttp.NewTransport(promTransport)
	transports[key] = t

	return t
}
