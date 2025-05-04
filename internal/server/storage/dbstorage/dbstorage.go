package dbstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"go.uber.org/zap"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(db *sql.DB) (*DBStorage, error) {
	query := `
	CREATE TABLE IF NOT EXISTS metrics (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		delta BIGINT,
		value DOUBLE PRECISION
	)`

	_, err := db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics table: %w", err)
	}

	return &DBStorage{db: db}, nil
}

func (s *DBStorage) UpdateMetric(name string, metric repositories.Metric) error {
	query := `
		INSERT INTO metrics (id, type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE 
		SET delta = EXCLUDED.delta, value = EXCLUDED.value`
	var delta sql.NullInt64
	var value sql.NullFloat64

	switch metric.Type {
	case "gauge":
		value = sql.NullFloat64{Float64: metric.Value.(float64), Valid: true}
		delta = sql.NullInt64{}
	case "counter":
		delta = sql.NullInt64{Int64: metric.Value.(int64), Valid: true}
		value = sql.NullFloat64{}
	default:
		return fmt.Errorf("unknown metric type: %s", metric.Type)
	}

	_, err := s.db.ExecContext(context.Background(), query, name, metric.Type, delta, value)
	return err
}

func (s *DBStorage) GetMetric(name string) (repositories.Metric, error) {
	query := `SELECT type, delta, value FROM metrics WHERE id = $1`
	row := s.db.QueryRowContext(context.Background(), query, name)

	var typ string
	var delta sql.NullInt64
	var value sql.NullFloat64

	err := row.Scan(&typ, &delta, &value)
	if err != nil {
		return repositories.Metric{}, err
	}

	var val interface{}
	switch typ {
	case "gauge":
		val = value.Float64
	case "counter":
		val = delta.Int64
	default:
		return repositories.Metric{}, fmt.Errorf("unknown type: %s", typ)
	}

	return repositories.Metric{Type: typ, Value: val}, nil
}

func (s *DBStorage) GetMetrics() (map[string]repositories.Metric, error) {
	query := `SELECT id, type, delta, value FROM metrics`
	rows, err := s.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Log.Error("Error closing rows", zap.Error(closeErr))
		}
	}()

	metrics := make(map[string]repositories.Metric)
	for rows.Next() {
		var id, metricType string
		var delta sql.NullInt64
		var value sql.NullFloat64

		if err := rows.Scan(&id, &metricType, &delta, &value); err != nil {
			return nil, err
		}

		var val interface{}
		switch metricType {
		case "gauge":
			val = value.Float64
		case "counter":
			val = delta.Int64
		default:
			return nil, errors.New("unknown metric type: " + metricType)
		}

		metrics[id] = repositories.Metric{Type: metricType, Value: val}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during rows iteration: %w", err)
	}

	return metrics, nil
}
