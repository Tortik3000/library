package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func init() {
	prometheus.Register(OutboxTasksCreated)
	prometheus.Register(OutboxTasksProcessed)
	prometheus.Register(OutboxTasksFailed)
	prometheus.Register(OutboxTaskProcessingDuration)

	OutboxTasksFailed.WithLabelValues("author").Add(0)
	OutboxTasksFailed.WithLabelValues("book").Add(0)
}

var (
	OutboxTasksCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "outbox",
		Name:      "tasks_created_total",
		Help:      "Общее число задач, созданных в outbox по each kind",
	}, []string{"kind"})

	OutboxTasksProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "outbox",
		Name:      "tasks_processed_total",
		Help:      "Общее число задач, успешно обработанных воркером, по each kind",
	}, []string{"kind"})

	OutboxTasksFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "outbox",
		Name:      "tasks_failed_total",
		Help:      "Общее число задач, завершившихся ошибкой, по each kind",
	}, []string{"kind"})

	OutboxTaskProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "outbox",
		Name:      "task_processing_duration_seconds",
		Help:      "Время обработки одной задачи из outbox (в секундах) по each kind",
		Buckets:   prometheus.DefBuckets,
	}, []string{"kind"})
)
