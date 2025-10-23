package messages

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "sms"
	metricsSubsystem = "messages"

	metricMessagesTotal = "total"
	metricCacheHits     = "cache_hits_total"
	metricCacheMisses   = "cache_misses_total"

	labelState = "state"
)

type metrics struct {
	totalCounter *prometheus.CounterVec

	cacheHits   prometheus.Counter
	cacheMisses prometheus.Counter
}

func newMetrics() *metrics {
	return &metrics{
		totalCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      metricMessagesTotal,
			Help:      "Total number of messages by state",
		}, []string{labelState}),

		cacheHits: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      metricCacheHits,
			Help:      "Number of cache hits",
		}),
		cacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      metricCacheMisses,
			Help:      "Number of cache misses",
		}),
	}
}

func (m *metrics) IncTotal(state string) {
	m.totalCounter.WithLabelValues(state).Inc()
}

func (m *metrics) IncCache(hit bool) {
	if hit {
		m.cacheHits.Inc()
	} else {
		m.cacheMisses.Inc()
	}
}
