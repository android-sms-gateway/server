package sse

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metric constants.
const (
	MetricActiveConnections = "active_connections"
	MetricEventsSent        = "events_sent_total"
	MetricConnectionErrors  = "connection_errors_total"
	MetricEventLatency      = "event_delivery_latency_seconds"
	MetricKeepalivesSent    = "keepalives_sent_total"

	LabelEventType = "event_type"
	LabelErrorType = "error_type"

	ErrorTypeBufferFull   = "buffer_full"
	ErrorTypeNoConnection = "no_connection"
	ErrorTypeWriteFailure = "write_failure"
	ErrorTypeMarshalError = "marshal_error"
)

// metrics contains all Prometheus metrics for the SSE module.
type metrics struct {
	activeConnections    *prometheus.GaugeVec
	eventsSent           *prometheus.CounterVec
	connectionErrors     *prometheus.CounterVec
	eventDeliveryLatency *prometheus.HistogramVec
	keepalivesSent       *prometheus.CounterVec
}

// newMetrics creates and initializes all SSE metrics.
func newMetrics() *metrics {
	var defBuckets = []float64{1e-6, 5e-6, 1e-5, 5e-5, 1e-4, 5e-4, .001, .005, .01, .05, .1}

	metrics := &metrics{
		activeConnections: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "sms",
			Subsystem: "sse",
			Name:      MetricActiveConnections,
			Help:      "Current number of active SSE connections",
		}, []string{}),
		eventsSent: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "sse",
			Name:      MetricEventsSent,
			Help:      "Total number of SSE events sent, labeled by event type",
		}, []string{LabelEventType}),
		connectionErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "sse",
			Name:      MetricConnectionErrors,
			Help:      "Total number of SSE connection errors, labeled by error type",
		}, []string{LabelErrorType}),
		eventDeliveryLatency: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "sms",
			Subsystem: "sse",
			Name:      MetricEventLatency,
			Help:      "Event delivery latency in seconds",
			Buckets:   defBuckets,
		}, []string{}),
		keepalivesSent: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "sse",
			Name:      MetricKeepalivesSent,
			Help:      "Total keepalive messages sent",
		}, []string{}),
	}

	return metrics
}

func (m *metrics) IncrementActiveConnections() {
	m.activeConnections.WithLabelValues().Inc()
}

func (m *metrics) DecrementActiveConnections() {
	m.activeConnections.WithLabelValues().Dec()
}

func (m *metrics) IncrementEventsSent(eventType string) {
	m.eventsSent.WithLabelValues(eventType).Inc()
}

func (m *metrics) IncrementConnectionErrors(errorType string) {
	m.connectionErrors.WithLabelValues(errorType).Inc()
}

func (m *metrics) ObserveEventDeliveryLatency(f func()) {
	timer := prometheus.NewTimer(m.eventDeliveryLatency.WithLabelValues())
	f()
	timer.ObserveDuration()
}

func (m *metrics) IncrementKeepalivesSent() {
	m.keepalivesSent.WithLabelValues().Inc()
}
