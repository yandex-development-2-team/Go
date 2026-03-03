package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	DBQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_database_queries_total",
			Help: "Total database queries",
		},
		[]string{"operation"},
	)

	DBErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_database_errors_total",
			Help: "Total database errors",
		},
		[]string{"operation"},
	)

	DBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bot_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(DBQueriesTotal, DBErrorsTotal, DBQueryDuration)
}
