package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.Register(GRPCRequestsTotal)
	prometheus.Register(GRPCRequestDuration)
}

var (
	GRPCRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "grpc",
		Name:      "requests_total",
		Help:      "Общее число gRPC-запросов",
	}, []string{"service", "method", "code"})

	GRPCRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "grpc",
		Name:      "request_duration_seconds",
		Help:      "Время обработки gRPC-запроса (в секундах)",
		Buckets:   prometheus.DefBuckets,
	}, []string{"service", "method"})
)
