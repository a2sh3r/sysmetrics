package restore

import (
	"context"
	"testing"

	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
)

func BenchmarkRestoreFromFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = RestoreFromFile("/tmp/metrics-db.json")
	}
}

func BenchmarkSaveToFile(b *testing.B) {
	repo := &mockStorage{}
	rc := NewRestoreConfig(1, "/tmp/metrics-db.json", repo)
	for i := 0; i < b.N; i++ {
		_ = rc.SaveToFile()
	}
}

type mockStorage struct{}

func (m *mockStorage) UpdateMetric(ctx context.Context, metricName string, metric repositories.Metric) error {
	return nil
}
func (m *mockStorage) GetMetric(ctx context.Context, metricName string) (repositories.Metric, error) {
	return repositories.Metric{}, nil
}
func (m *mockStorage) GetMetrics(ctx context.Context) (map[string]repositories.Metric, error) {
	return map[string]repositories.Metric{}, nil
}
func (m *mockStorage) UpdateMetricsBatch(ctx context.Context, metrics map[string]repositories.Metric) error {
	return nil
}
