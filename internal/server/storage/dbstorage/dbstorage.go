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
	switch metric.Type {
	case "gauge":
		query := `
			INSERT INTO metrics (id, type, delta, value)
			VALUES ($1, 'gauge', NULL, $2)
			ON CONFLICT (id) DO UPDATE 
			SET delta = NULL,
				value = $2`
		value := metric.Value.(float64)
		_, err := s.db.ExecContext(context.Background(), query, name, value)
		return err
	case "counter":
		query := `
			INSERT INTO metrics (id, type, delta, value)
			VALUES ($1, 'counter', $2, NULL)
			ON CONFLICT (id) DO UPDATE 
			SET delta = metrics.delta + $2,
				value = NULL`
		delta := metric.Value.(int64)
		_, err := s.db.ExecContext(context.Background(), query, name, delta)
		return err
	default:
		return fmt.Errorf("unknown metric type: %s", metric.Type)
	}
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

func (s *DBStorage) UpdateMetricsBatch(metrics map[string]repositories.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Log.Error("Failed to rollback transaction", zap.Error(rollbackErr))
			}
		}
	}()

	counterQuery := `
		INSERT INTO metrics (id, type, delta, value)
		VALUES ($1, 'counter', $2, NULL)
		ON CONFLICT (id) DO UPDATE 
		SET delta = metrics.delta + $2,
		value = NULL`

	gaugeQuery := `
		INSERT INTO metrics (id, type, delta, value)
		VALUES ($1, 'gauge', NULL, $2)
		ON CONFLICT (id) DO UPDATE 
		SET delta = NULL,
			value = $2`

	counterStmt, err := tx.PrepareContext(context.Background(), counterQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare counter statement: %w", err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			logger.Log.Error("unable to close counter stmt", zap.Error(err))
		}
	}(counterStmt)

	gaugeStmt, err := tx.PrepareContext(context.Background(), gaugeQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare gauge statement: %w", err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			logger.Log.Error("unable to close gauge stmt", zap.Error(err))
		}
	}(gaugeStmt)

	for id, metric := range metrics {
		switch metric.Type {
		case "gauge":
			value := metric.Value.(float64)
			if _, err := gaugeStmt.ExecContext(context.Background(), id, value); err != nil {
				return fmt.Errorf("failed to execute gauge statement for metric %s: %w", id, err)
			}
		case "counter":
			delta := metric.Value.(int64)
			if _, err := counterStmt.ExecContext(context.Background(), id, delta); err != nil {
				return fmt.Errorf("failed to execute counter statement for metric %s: %w", id, err)
			}
		default:
			return fmt.Errorf("unknown metric type: %s", metric.Type)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
