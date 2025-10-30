package push

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type BlacklistOperation string

const (
	BlacklistOperationAdded   BlacklistOperation = "added"
	BlacklistOperationSkipped BlacklistOperation = "skipped"
)

type metrics struct {
	enqueuedCounter  *prometheus.CounterVec
	retriesCounter   prometheus.Counter
	blacklistCounter *prometheus.CounterVec
	errorsCounter    *prometheus.CounterVec
}

func newMetrics() *metrics {
	return &metrics{
		enqueuedCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "push",
			Name:      "enqueued_total",
			Help:      "Total number of messages enqueued",
		}, []string{"event"}),

		retriesCounter: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "push",
			Name:      "retries_total",
			Help:      "Total retry attempts",
		}),

		blacklistCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "push",
			Name:      "blacklist_total",
			Help:      "Blacklist operations",
		}, []string{"operation"}),

		errorsCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "push",
			Name:      "errors_total",
			Help:      "Total number of errors",
		}, []string{}),
	}
}

func (m *metrics) IncEnqueued(event string) {
	m.enqueuedCounter.WithLabelValues(event).Inc()
}

func (m *metrics) IncRetry() {
	m.retriesCounter.Inc()
}

func (m *metrics) IncBlacklist(operation BlacklistOperation) {
	m.blacklistCounter.WithLabelValues(string(operation)).Inc()
}

func (m *metrics) IncError(v int) {
	m.errorsCounter.WithLabelValues().Add(float64(v))
}
