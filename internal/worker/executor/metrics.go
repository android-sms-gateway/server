package executor

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metricsTaskResult string

const (
	metricsTaskResultSuccess metricsTaskResult = "success"
	metricsTaskResultError   metricsTaskResult = "error"
)

type metrics struct {
	activeTasksCounter prometheus.Gauge
	taskResult         *prometheus.CounterVec
	taskDuration       *prometheus.HistogramVec
}

func newMetrics() *metrics {
	var defBuckets = []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 25}

	return &metrics{
		activeTasksCounter: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "worker",
			Subsystem: "executor",
			Name:      "active_tasks",
			Help:      "Number of active tasks",
		}),
		taskResult: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "worker",
			Subsystem: "executor",
			Name:      "task_result_total",
			Help:      "Task result, labeled by task name and result",
		}, []string{"task", "result"}),
		taskDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "worker",
			Subsystem: "executor",
			Name:      "task_duration_seconds",
			Help:      "Task duration in seconds",
			Buckets:   defBuckets,
		}, []string{"task"}),
	}
}

func (m *metrics) IncActiveTasks() {
	m.activeTasksCounter.Inc()
}

func (m *metrics) DecActiveTasks() {
	m.activeTasksCounter.Dec()
}

func (m *metrics) ObserveTaskResult(task string, result metricsTaskResult, duration time.Duration) {
	m.taskResult.WithLabelValues(task, string(result)).Inc()
	m.taskDuration.WithLabelValues(task).Observe(duration.Seconds())
}
