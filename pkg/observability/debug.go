package observability

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewDebugServer serves operational endpoints on a port separate from user traffic:
// pprof at /debug/pprof, expvar at /debug/vars, the process Prometheus registry at /metrics and probes at /liveness and /readiness
func NewDebugServer(address string, liveness, readiness http.HandlerFunc) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("GET /debug/vars", expvar.Handler())
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /liveness", liveness)
	mux.HandleFunc("GET /readiness", readiness)
	return &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       5 * time.Second,
	}
}
