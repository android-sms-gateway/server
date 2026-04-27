package messages

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricsNamespace = "sms"
	metricsSubsystem = "messages"

	metricMessagesTotal           = "total"
	metricCacheHits               = "cache_hits_total"
	metricCacheMisses             = "cache_misses_total"
	metricLimiterChecksTotal      = "limiter_checks_total"
	metricLimiterRefreshesTotal   = "limiter_refreshes_total"
	metricLimiterBatchSize        = "limiter_batch_size"
	metricLimiterQueryErrorsTotal = "limiter_query_errors_total"

	labelState  = "state"
	labelResult = "result"
	labelCheck  = "check"

	resultAllowed = "allowed"
	resultLimited = "limited"

	checkMaxPending    = "max_pending"
	checkMaxPendingAge = "max_pending_age"
	checkMaxFailed     = "max_failed"
)

type metrics struct {
	totalCounter *prometheus.CounterVec

	cacheHits   prometheus.Counter
	cacheMisses prometheus.Counter

	limiterChecksTotal    *prometheus.CounterVec
	limiterRefreshesTotal prometheus.Counter
	limiterBatchSize      prometheus.Gauge
	limiterQueryErrors    *prometheus.CounterVec
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

		limiterChecksTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      metricLimiterChecksTotal,
			Help:      "Total number of limiter checks by result",
		}, []string{labelResult}),

		limiterRefreshesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      metricLimiterRefreshesTotal,
			Help:      "Total number of limiter refresh requests",
		}),

		limiterBatchSize: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      metricLimiterBatchSize,
			Help:      "Number of devices processed per refresh cycle",
		}),

		limiterQueryErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: metricsSubsystem,
			Name:      metricLimiterQueryErrorsTotal,
			Help:      "Total number of limiter query errors by check type",
		}, []string{labelCheck}),
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

func (m *metrics) IncLimiterCheck(limited bool) {
	result := resultAllowed
	if limited {
		result = resultLimited
	}
	m.limiterChecksTotal.WithLabelValues(result).Inc()
}

func (m *metrics) IncLimiterRefresh() {
	m.limiterRefreshesTotal.Inc()
}

func (m *metrics) SetLimiterBatchSize(size int) {
	m.limiterBatchSize.Set(float64(size))
}

func (m *metrics) IncLimiterQueryError(check string) {
	m.limiterQueryErrors.WithLabelValues(check).Inc()
}
