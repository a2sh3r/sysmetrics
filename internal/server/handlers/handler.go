package handlers

import (
	"database/sql"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
)

type ReaderServiceInterface interface {
	GetMetric(metricName string) (repositories.Metric, error)
	GetMetrics() (map[string]repositories.Metric, error)
}

type WriterServiceInterface interface {
	UpdateGaugeMetric(name string, value float64) error
	UpdateCounterMetric(name string, value int64) error
	UpdateMetricsBatch(metrics map[string]repositories.Metric) error
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
