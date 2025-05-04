package handlers

import (
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
)

type ReaderServiceInterface interface {
	GetMetric(metricName string) (repositories.Metric, error)
	GetMetrics() (map[string]repositories.Metric, error)
}

type WriterServiceInterface interface {
	UpdateGaugeMetric(name string, value float64) error
	UpdateCounterMetric(name string, value int64) error
}

type Handler struct {
	reader ReaderServiceInterface
	writer WriterServiceInterface
}

func NewHandler(reader ReaderServiceInterface, writer WriterServiceInterface) *Handler {
	return &Handler{
		reader: reader,
		writer: writer,
	}
}
