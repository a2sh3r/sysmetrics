package metric

import (
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
	"testing"
)

func TestService_UpdateGaugeMetric(t *testing.T) {
	storage := memstorage.NewMemStorage()
	metricRepo := repositories.NewMetricRepo(storage)
	service := NewService(metricRepo)

	err := service.UpdateGaugeMetric("test_metric", 10.5)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestService_UpdateCounterMetric(t *testing.T) {
	storage := memstorage.NewMemStorage()
	metricRepo := repositories.NewMetricRepo(storage)
	service := NewService(metricRepo)

	err := service.UpdateCounterMetric("test_metric", 10)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestService_UpdateMetricNameEmpty(t *testing.T) {
	storage := memstorage.NewMemStorage()
	metricRepo := repositories.NewMetricRepo(storage)
	service := NewService(metricRepo)

	err := service.UpdateGaugeMetric("", 10.5)
	if err == nil {
		t.Errorf("Expected error for empty metric name, but got nil")
	}
}
