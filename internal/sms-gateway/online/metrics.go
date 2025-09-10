package online

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metric constants
const (
	metricStatusSetTotal     = "status_set_total"
	metricCacheOperations    = "cache_operations_total"
	metricCacheLatency       = "cache_latency_seconds"
	metricPersistenceLatency = "persistence_latency_seconds"
	metricPersistenceErrors  = "persistence_errors_total"
	metricBatchSize          = "batch_size"

	labelOperation = "operation"
	labelStatus    = "status"

	operationSet   = "set"
	operationDrain = "drain"

	statusSuccess = "success"
	statusError   = "error"
)

// metrics contains all Prometheus metrics for the online module
type metrics struct {
	statusSetCounter   *prometheus.CounterVec
	cacheOperations    *prometheus.CounterVec
	cacheLatency       prometheus.Histogram
	persistenceLatency prometheus.Histogram
	persistenceErrors  prometheus.Counter
	batchSize          prometheus.Gauge
}

// newMetrics creates and initializes all online metrics
func newMetrics() *metrics {
	return &metrics{
		statusSetCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "online",
			Name:      metricStatusSetTotal,
			Help:      "Total number of online status updates",
		}, []string{labelStatus}),

		cacheOperations: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "online",
			Name:      metricCacheOperations,
			Help:      "Total cache operations by type",
		}, []string{labelOperation, labelStatus}),

		cacheLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "sms",
			Subsystem: "online",
			Name:      metricCacheLatency,
			Help:      "Cache operation latency in seconds",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		}),

		persistenceLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "sms",
			Subsystem: "online",
			Name:      metricPersistenceLatency,
			Help:      "Persistence operation latency in seconds",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		}),

		persistenceErrors: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "online",
			Name:      metricPersistenceErrors,
			Help:      "Total persistence errors by type",
		}),

		batchSize: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "sms",
			Subsystem: "online",
			Name:      metricBatchSize,
			Help:      "Current batch size",
		}),
	}
}

// IncrementStatusSet increments the status set counter
func (m *metrics) IncrementStatusSet(success bool) {
	status := statusSuccess
	if !success {
		status = statusError
	}
	m.statusSetCounter.WithLabelValues(status).Inc()
}

// IncrementCacheOperation increments cache operation counter
func (m *metrics) IncrementCacheOperation(operation, status string) {
	m.cacheOperations.WithLabelValues(operation, status).Inc()
}

// ObserveCacheLatency observes cache operation latency
func (m *metrics) ObserveCacheLatency(f func()) {
	timer := prometheus.NewTimer(m.cacheLatency)
	f()
	timer.ObserveDuration()
}

// ObservePersistenceLatency observes persistence operation latency
func (m *metrics) ObservePersistenceLatency(f func()) {
	timer := prometheus.NewTimer(m.persistenceLatency)
	f()
	timer.ObserveDuration()
}

// IncrementPersistenceError increments persistence error counter
func (m *metrics) IncrementPersistenceError() {
	m.persistenceErrors.Inc()
}

// SetBatchSize sets the current batch size
func (m *metrics) SetBatchSize(size int) {
	m.batchSize.Set(float64(size))
}
