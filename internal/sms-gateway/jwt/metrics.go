package jwt

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metric constants.
const (
	MetricTokensIssuedTotal         = "jwt_tokens_issued_total"    //nolint:gosec // false positive
	MetricTokensValidatedTotal      = "jwt_tokens_validated_total" //nolint:gosec // false positive
	MetricTokensRevokedTotal        = "jwt_tokens_revoked_total"   //nolint:gosec // false positive
	MetricIssuanceDurationSeconds   = "jwt_issuance_duration_seconds"
	MetricValidationDurationSeconds = "jwt_validation_duration_seconds"
	MetricRevocationDurationSeconds = "jwt_revocation_duration_seconds"

	labelStatus = "status"

	StatusSuccess = "success"
	StatusError   = "error"
)

// Metrics contains all Prometheus Metrics for the JWT module.
type Metrics struct {
	tokensIssuedCounter         *prometheus.CounterVec
	tokensValidatedCounter      *prometheus.CounterVec
	tokensRevokedCounter        *prometheus.CounterVec
	issuanceDurationHistogram   prometheus.Histogram
	validationDurationHistogram prometheus.Histogram
	revocationDurationHistogram prometheus.Histogram
}

// NewMetrics creates and initializes all JWT metrics.
func NewMetrics() *Metrics {
	var defBuckets = []float64{.0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1}
	return &Metrics{
		tokensIssuedCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "auth",
			Name:      MetricTokensIssuedTotal,
			Help:      "Total number of JWT tokens issued",
		}, []string{labelStatus}),

		tokensValidatedCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "auth",
			Name:      MetricTokensValidatedTotal,
			Help:      "Total number of JWT tokens validated",
		}, []string{labelStatus}),

		tokensRevokedCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sms",
			Subsystem: "auth",
			Name:      MetricTokensRevokedTotal,
			Help:      "Total number of JWT tokens revoked",
		}, []string{labelStatus}),

		issuanceDurationHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "sms",
			Subsystem: "auth",
			Name:      MetricIssuanceDurationSeconds,
			Help:      "JWT issuance duration in seconds",
			Buckets:   defBuckets,
		}),

		validationDurationHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "sms",
			Subsystem: "auth",
			Name:      MetricValidationDurationSeconds,
			Help:      "JWT validation duration in seconds",
			Buckets:   defBuckets,
		}),

		revocationDurationHistogram: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "sms",
			Subsystem: "auth",
			Name:      MetricRevocationDurationSeconds,
			Help:      "JWT revocation duration in seconds",
			Buckets:   defBuckets,
		}),
	}
}

// IncrementTokensIssued increments the tokens issued counter.
func (m *Metrics) IncrementTokensIssued(status string) {
	m.tokensIssuedCounter.WithLabelValues(status).Inc()
}

// IncrementTokensValidated increments the tokens validated counter.
func (m *Metrics) IncrementTokensValidated(status string) {
	m.tokensValidatedCounter.WithLabelValues(status).Inc()
}

// IncrementTokensRevoked increments the tokens revoked counter.
func (m *Metrics) IncrementTokensRevoked(status string) {
	m.tokensRevokedCounter.WithLabelValues(status).Inc()
}

// ObserveIssuance observes issuance duration.
func (m *Metrics) ObserveIssuance(f func()) {
	timer := prometheus.NewTimer(m.issuanceDurationHistogram)
	defer timer.ObserveDuration()
	f()
}

// ObserveValidation observes validation duration.
func (m *Metrics) ObserveValidation(f func()) {
	timer := prometheus.NewTimer(m.validationDurationHistogram)
	defer timer.ObserveDuration()
	f()
}

// ObserveRevocation observes revocation duration.
func (m *Metrics) ObserveRevocation(f func()) {
	timer := prometheus.NewTimer(m.revocationDurationHistogram)
	defer timer.ObserveDuration()
	f()
}
