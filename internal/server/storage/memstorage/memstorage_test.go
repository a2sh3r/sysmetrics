package memstorage

import (
	"errors"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	ms := NewMemStorage()
	if ms == nil {
		t.Errorf("NewMemStorage returned nil")
	}
	if ms != nil {
		if ms.metrics == nil {
			t.Errorf("Metrics map is nil")
		}
	}

}

func TestGetMetric(t *testing.T) {
	ms := NewMemStorage()
	_ = ms.UpdateMetric("test", repositories.Metric{Type: "gauge", Value: 156.14})

	metric, err := ms.GetMetric("test")
	if err != nil {
		t.Errorf("GetMetric returned error: %v", err)
	}
	if metric.Type != "gauge" || metric.Value != 156.14 {
		t.Errorf("GetMetric returned incorrect metric: %+v", metric)
	}

	_, err = ms.GetMetric("testers")
	if !errors.Is(err, ErrMetricNotFound) {
		t.Errorf("GetMetric did not return ErrMetricNotFound for non-existent metric")
	}
}

func TestUpdateMetric_NewCounter(t *testing.T) {
	ms := NewMemStorage()

	err := ms.UpdateMetric("test_int64", repositories.Metric{Type: "counter", Value: int64(10)})
	if err != nil {
		t.Fatalf("UpdateMetric failed: %v", err)
	}

	m, err := ms.GetMetric("test_int64")
	if err != nil {
		t.Fatalf("GetMetric failed: %v", err)
	}

	if m.Value.(int64) != int64(10) {
		t.Fatalf("Expected int64 value 10, got %v", m.Value)
	}
}

func TestUpdateMetric_UpdateCounter(t *testing.T) {
	ms := NewMemStorage()

	err := ms.UpdateMetric("test_int64", repositories.Metric{Type: "counter", Value: int64(10)})
	if err != nil {
		t.Fatalf("UpdateMetric failed: %v", err)
	}

	err = ms.UpdateMetric("test_int64", repositories.Metric{Type: "counter", Value: int64(5)})
	if err != nil {
		t.Fatalf("UpdateMetric failed: %v", err)
	}

	m, err := ms.GetMetric("test_int64")
	if err != nil {
		t.Fatalf("GetMetric failed: %v", err)
	}

	if m.Value.(int64) != int64(15) {
		t.Fatalf("Expected int64 value 15, got %v", m.Value)
	}
}

func TestUpdateMetric_NewGauge(t *testing.T) {
	ms := NewMemStorage()

	err := ms.UpdateMetric("test_float64", repositories.Metric{Type: "gauge", Value: 3.14})
	if err != nil {
		t.Fatalf("UpdateMetric failed: %v", err)
	}

	m, err := ms.GetMetric("test_float64")
	if err != nil {
		t.Fatalf("GetMetric failed: %v", err)
	}

	if m.Value.(float64) != float64(3.14) {
		t.Fatalf("Expected float64 value 3.14, got %v", m.Value)
	}
}

func TestUpdateMetric_UpdateGauge(t *testing.T) {
	ms := NewMemStorage()

	err := ms.UpdateMetric("test_float64", repositories.Metric{Type: "gauge", Value: 3.14})
	if err != nil {
		t.Fatalf("UpdateMetric failed: %v", err)
	}

	err = ms.UpdateMetric("test_float64", repositories.Metric{Type: "gauge", Value: 2.71})
	if err != nil {
		t.Fatalf("UpdateMetric failed: %v", err)
	}

	m, err := ms.GetMetric("test_float64")
	if err != nil {
		t.Fatalf("GetMetric failed: %v", err)
	}

	if m.Value.(float64) != float64(2.71) {
		t.Fatalf("Expected float64 value 2.71, got %v", m.Value)
	}
}

func TestUpdateMetric_InvalidType(t *testing.T) {
	ms := NewMemStorage()

	err := ms.UpdateMetric("test_invalid", repositories.Metric{Type: "invalid", Value: "some_value"})
	if err == nil {
		t.Fatal("Expected error for invalid metric type, got nil")
	}

	if !errors.Is(err, ErrMetricInvalidType) {
		t.Fatalf("Expected ErrMetricTypeInvalid, got %v", err)
	}
}

func TestUpdateMetric_NilStorage(t *testing.T) {
	var ms *MemStorage = nil

	err := ms.UpdateMetric("test_nil", repositories.Metric{Type: "counter", Value: int64(10)})
	if err == nil {
		t.Fatal("Expected error for nil storage, got nil")
	}

	if !errors.Is(err, ErrStorageNil) {
		t.Fatalf("Expected ErrStorageNil, got %v", err)
	}
}

func TestUpdateMetric_NilMetricsMap(t *testing.T) {
	ms := NewMemStorage()
	ms.metrics = nil

	err := ms.UpdateMetric("test_nil_map", repositories.Metric{Type: "counter", Value: int64(10)})
	if err == nil {
		t.Fatal("Expected error for nil metrics map, got nil")
	}

	if !errors.Is(err, ErrMetricsMapNil) {
		t.Fatalf("Expected ErrMetricsMapNil, got %v", err)
	}
}
