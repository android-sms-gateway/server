package messages

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "sms"
	metricsSubsystem = "messages"

	metricMessagesTotal = "total"

	labelState = "state"
)

type metrics struct {
	totalCounter *prometheus.CounterVec
}

func newMetrics() *metrics {
	return &metrics{
		totalCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      metricMessagesTotal,
			Help:      "Total number of messages by state",
		}, []string{labelState}),
	}
}

func (m *metrics) IncTotal(state string) {
	m.totalCounter.WithLabelValues(state).Inc()
}
