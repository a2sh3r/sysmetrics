package handlers

import (
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
)

type ServiceInterface interface {
	UpdateGaugeMetric(name string, value float64) error
	UpdateCounterMetric(name string, value int64) error
	GetMetric(metricName string) (repositories.Metric, error)
	GetMetrics() (map[string]repositories.Metric, error)
}

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}
