package dbstorage

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func BenchmarkUpdateMetric(b *testing.B) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		b.Skip("TEST_DATABASE_DSN not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		b.Fatalf("failed to open db: %v", err)
	}
	s, err := NewDBStorage(db)
	if err != nil {
		b.Fatalf("failed to create storage: %v", err)
	}
	ctx := context.Background()
	metric := repositories.Metric{Type: "gauge", Value: float64(42)}
	for i := 0; i < b.N; i++ {
		_ = s.UpdateMetric(ctx, "test_metric", metric)
	}
}
