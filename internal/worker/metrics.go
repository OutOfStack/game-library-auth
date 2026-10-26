package worker

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	pending   prometheus.Gauge
	oldest    prometheus.Gauge
	outcomes  *prometheus.CounterVec
	reclaimed prometheus.Counter
}

func newMetrics(registry prometheus.Registerer) (*metrics, error) {
	m := &metrics{
		pending:   prometheus.NewGauge(prometheus.GaugeOpts{Name: "notification_outbox_pending", Help: "Notifications awaiting completion."}),
		oldest:    prometheus.NewGauge(prometheus.GaugeOpts{Name: "notification_outbox_oldest_age_seconds", Help: "Age of the oldest pending notification."}),
		outcomes:  prometheus.NewCounterVec(prometheus.CounterOpts{Name: "notification_outbox_outcomes_total", Help: "Notification processing outcomes."}, []string{"outcome"}),
		reclaimed: prometheus.NewCounter(prometheus.CounterOpts{Name: "notification_outbox_reclaimed_total", Help: "Expired notification leases reclaimed."}),
	}
	for _, collector := range []prometheus.Collector{m.pending, m.oldest, m.outcomes, m.reclaimed} {
		if err := registry.Register(collector); err != nil {
			return nil, err
		}
	}
	return m, nil
}
