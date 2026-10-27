package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	pending   prometheus.Gauge
	oldest    prometheus.Gauge
	outcomes  *prometheus.CounterVec
	reclaimed prometheus.Counter
}

func newMetrics(registry prometheus.Registerer) *metrics {
	f := promauto.With(registry)
	return &metrics{
		pending:   f.NewGauge(prometheus.GaugeOpts{Name: "notification_outbox_pending", Help: "Notifications awaiting completion."}),
		oldest:    f.NewGauge(prometheus.GaugeOpts{Name: "notification_outbox_oldest_age_seconds", Help: "Age of the oldest pending notification."}),
		outcomes:  f.NewCounterVec(prometheus.CounterOpts{Name: "notification_outbox_outcomes_total", Help: "Notification processing outcomes."}, []string{"outcome"}),
		reclaimed: f.NewCounter(prometheus.CounterOpts{Name: "notification_outbox_reclaimed_total", Help: "Expired notification leases reclaimed."}),
	}
}
