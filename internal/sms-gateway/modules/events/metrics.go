package events

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metric constants
const (
	MetricEnqueuedTotal = "enqueued_total"
	MetricSentTotal     = "sent_total"
	MetricFailedTotal   = "failed_total"

	LabelEvent        = "event"
	LabelDeliveryType = "delivery_type"
	LabelReason       = "reason"

	DeliveryTypePush    = "push"
	DeliveryTypeSSE     = "sse"
	DeliveryTypeUnknown = "unknown"

	FailureReasonQueueFull      = "queue_full"
	FailureReasonProviderFailed = "provider_failed"
)

// metrics contains all Prometheus metrics for the events module
type metrics struct {
	enqueuedCounter *prometheus.CounterVec
	sentCounter     *prometheus.CounterVec
	failedCounter   *prometheus.CounterVec
}

// newMetrics creates and initializes all events metrics
func newMetrics() *metrics {
	return &metrics{
		enqueuedCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "events",
			Name:      MetricEnqueuedTotal,
			Help:      "Total number of events enqueued",
		}, []string{LabelEvent}),
		sentCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "events",
			Name:      MetricSentTotal,
			Help:      "Total number of events sent",
		}, []string{LabelEvent, LabelDeliveryType}),
		failedCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "events",
			Name:      MetricFailedTotal,
			Help:      "Total number of failed notifications",
		}, []string{LabelEvent, LabelDeliveryType, LabelReason}),
	}
}

// IncrementEnqueued increments the enqueued counter for the given event type
func (m *metrics) IncrementEnqueued(eventType string) {
	m.enqueuedCounter.WithLabelValues(eventType).Inc()
}

// IncrementSent increments the sent counter for the given event type and delivery type
func (m *metrics) IncrementSent(eventType string, deliveryType string) {
	m.sentCounter.WithLabelValues(eventType, deliveryType).Inc()
}

// IncrementFailed increments the failed counter for the given event type, delivery type, and reason
func (m *metrics) IncrementFailed(eventType string, deliveryType string, reason string) {
	m.failedCounter.WithLabelValues(eventType, deliveryType, reason).Inc()
}
