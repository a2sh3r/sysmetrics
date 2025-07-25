// Package handlers provides HTTP handlers for metrics operations.
package handlers

import (
	"context"
	"database/sql"

	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
)

// ReaderServiceInterface defines methods for reading metrics with retry logic.
type ReaderServiceInterface interface {
	GetMetricWithRetry(ctx context.Context, metricName string) (repositories.Metric, error)
	GetMetricsWithRetry(ctx context.Context) (map[string]repositories.Metric, error)
}

// WriterServiceInterface defines methods for updating metrics with retry logic.
type WriterServiceInterface interface {
	UpdateGaugeMetricWithRetry(ctx context.Context, name string, value float64) error
	UpdateCounterMetricWithRetry(ctx context.Context, name string, value int64) error
	UpdateMetricsBatchWithRetry(ctx context.Context, metrics map[string]repositories.Metric) error
}

// Handler handles HTTP requests for metrics.
type Handler struct {
	reader ReaderServiceInterface
	writer WriterServiceInterface
	DB     *sql.DB
}

// NewHandler creates a new Handler instance.
func NewHandler(reader ReaderServiceInterface, writer WriterServiceInterface, db *sql.DB) *Handler {
	return &Handler{
		reader: reader,
		writer: writer,
		DB:     db,
	}
}
