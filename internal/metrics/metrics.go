package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Metrics хранит все кастомные метрики приложения
type Metrics struct {
	// Сообщения от бота
	MessagesReceived          prometheus.Counter
	MessagesProcessedTotal    prometheus.Counter
	MessagesErrorsTotal       prometheus.Counter
	MessageProcessingDuration prometheus.Histogram

	// База данных
	DatabaseQueriesTotal  prometheus.Counter
	DatabaseQueryDuration prometheus.Histogram

	// Telegram API
	APIRequestsTotal prometheus.Counter

	// Бизнес-метрики
	ActiveUsers   prometheus.Gauge
	BookingsTotal prometheus.Counter

	registry *prometheus.Registry
	logger   *zap.Logger
}

// NewMetrics создает и регистрирует все кастомные метрики
func NewMetrics(logger *zap.Logger) (*Metrics, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	registry := prometheus.NewRegistry()

	m := &Metrics{
		registry: registry,
		logger:   logger,
	}

	// Counter: всего сообщений получено
	m.MessagesReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "bot",
		Name:      "messages_received_total",
		Help:      "Total number of messages received from users",
	})

	// Counter: всего сообщений обработано
	m.MessagesProcessedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "bot",
		Name:      "messages_processed_total",
		Help:      "Total number of messages successfully processed",
	})

	// Counter: ошибки при обработке
	m.MessagesErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "bot",
		Name:      "messages_errors_total",
		Help:      "Total number of message processing errors",
	})

	// Histogram: время обработки сообщений
	m.MessageProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "bot",
		Name:      "message_processing_duration_seconds",
		Help:      "Time spent processing messages in seconds",
		Buckets:   prometheus.DefBuckets,
	})

	// Counter: всего запросов к БД
	m.DatabaseQueriesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "bot",
		Name:      "database_queries_total",
		Help:      "Total number of database queries",
	})

	// Histogram: время запросов к БД
	m.DatabaseQueryDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "bot",
		Name:      "database_query_duration_seconds",
		Help:      "Time spent on database queries in seconds",
		Buckets:   prometheus.DefBuckets,
	})

	// Counter: запросы к Telegram API
	m.APIRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "bot",
		Name:      "api_requests_total",
		Help:      "Total number of requests to Telegram API",
	})

	// Gauge: количество активных пользователей
	m.ActiveUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "bot",
		Name:      "active_users",
		Help:      "Number of active users",
	})

	// Counter: всего бронирований
	m.BookingsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "bot",
		Name:      "bookings_total",
		Help:      "Total number of bookings made",
	})

	// Регистрируем все метрики
	collectors := []prometheus.Collector{
		m.MessagesReceived,
		m.MessagesProcessedTotal,
		m.MessagesErrorsTotal,
		m.MessageProcessingDuration,
		m.DatabaseQueriesTotal,
		m.DatabaseQueryDuration,
		m.APIRequestsTotal,
		m.ActiveUsers,
		m.BookingsTotal,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// Collector возвращает registry для интеграции с сервером
func (m *Metrics) Collector() *prometheus.Registry {
	return m.registry
}
