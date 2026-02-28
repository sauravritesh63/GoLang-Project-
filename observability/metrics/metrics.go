// Package metrics exposes Prometheus metrics for the distributed task scheduler.
// Register counters, histograms, and gauges here; then call Register() once
// during application startup to add them to the default Prometheus registry.
//
// Exposed metrics:
//
//	scheduler_workflows_total           – total workflow runs triggered (labels: status)
//	scheduler_tasks_total               – total task runs processed  (labels: status)
//	scheduler_task_duration_seconds     – task execution duration histogram
//	scheduler_workflow_failures_total   – total workflow run failures
//	scheduler_workflow_successes_total  – total workflow run successes
//	scheduler_worker_heartbeats_total   – total worker heartbeat ticks (labels: worker_id)
//	scheduler_task_retries_total        – total task retry attempts   (labels: worker_id)
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collector groups all Prometheus metrics exposed by the scheduler system.
type Collector struct {
	WorkflowsTotal   *prometheus.CounterVec
	TasksTotal       *prometheus.CounterVec
	TaskDuration     *prometheus.HistogramVec
	WorkflowFailures prometheus.Counter
	WorkflowSuccesses prometheus.Counter
	WorkerHeartbeats *prometheus.CounterVec
	TaskRetries      *prometheus.CounterVec
}

// New registers and returns all scheduler Prometheus metrics using promauto so
// that each metric is automatically registered with the default registry.
func New() *Collector {
	return &Collector{
		WorkflowsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "scheduler_workflows_total",
			Help: "Total number of workflow runs triggered.",
		}, []string{"status"}),

		TasksTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "scheduler_tasks_total",
			Help: "Total number of task runs processed.",
		}, []string{"status"}),

		TaskDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "scheduler_task_duration_seconds",
			Help:    "Histogram of task execution durations in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"status"}),

		WorkflowFailures: promauto.NewCounter(prometheus.CounterOpts{
			Name: "scheduler_workflow_failures_total",
			Help: "Total number of workflow run failures.",
		}),

		WorkflowSuccesses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "scheduler_workflow_successes_total",
			Help: "Total number of workflow run successes.",
		}),

		WorkerHeartbeats: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "scheduler_worker_heartbeats_total",
			Help: "Total number of worker heartbeat ticks.",
		}, []string{"worker_id"}),

		TaskRetries: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "scheduler_task_retries_total",
			Help: "Total number of task retry attempts.",
		}, []string{"worker_id"}),
	}
}
