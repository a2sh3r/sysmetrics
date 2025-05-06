package handlers

import (
	"context"
	"database/sql"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
)

type ReaderServiceInterface interface {
	GetMetric(ctx context.Context, metricName string) (repositories.Metric, error)
	GetMetricWithRetry(ctx context.Context, metricName string) (repositories.Metric, error)
	GetMetrics(ctx context.Context) (map[string]repositories.Metric, error)
	GetMetricsWithRetry(ctx context.Context) (map[string]repositories.Metric, error)
}

type WriterServiceInterface interface {
	UpdateGaugeMetric(ctx context.Context, name string, value float64) error
	UpdateCounterMetric(ctx context.Context, name string, value int64) error
	UpdateMetricsBatch(ctx context.Context, metrics map[string]repositories.Metric) error
	UpdateGaugeMetricWithRetry(ctx context.Context, name string, value float64) error
	UpdateCounterMetricWithRetry(ctx context.Context, name string, value int64) error
	UpdateMetricsBatchWithRetry(ctx context.Context, metrics map[string]repositories.Metric) error
}

type Handler struct {
	reader ReaderServiceInterface
	writer WriterServiceInterface
	DB     *sql.DB
}

func NewHandler(reader ReaderServiceInterface, writer WriterServiceInterface, db *sql.DB) *Handler {
	return &Handler{
		reader: reader,
		writer: writer,
		DB:     db,
	}
}
