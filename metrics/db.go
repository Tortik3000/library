package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.Register(DBTableInsertRate)
	prometheus.Register(DBTableRowsCount)
	prometheus.Register(DBQueryLatency)

}

var (
	DBTableRowsCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "db",
		Name:      "table_num_rows",
		Help:      "Количество записей в таблицах базы",
	}, []string{"table"})

	DBTableInsertRate = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "db",
		Name:      "table_insert_total",
		Help:      "Общее число вставок в каждую таблицу",
	}, []string{"table"})

	DBQueryLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_latency_seconds",
			Help:    "Latency of DB queries by operation",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)
