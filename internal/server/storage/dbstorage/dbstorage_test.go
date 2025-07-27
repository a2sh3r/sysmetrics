package dbstorage_test

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/dbstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func expectTableCreation(mock sqlmock.Sqlmock) {
	mock.ExpectExec(regexp.QuoteMeta(`
    CREATE TABLE IF NOT EXISTS metrics (
        id TEXT PRIMARY KEY,
        type TEXT NOT NULL,
        delta BIGINT,
        value DOUBLE PRECISION
    )
    `)).WillReturnResult(sqlmock.NewResult(0, 0))
}

func TestDBStorage_UpdateMetric(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if errDB := db.Close(); errDB != nil {
			fmt.Printf("error closing db")
		}
	}()
	expectTableCreation(mock)
	storage, err := dbstorage.NewDBStorage(db)
	require.NoError(t, err)

	tests := []struct {
		name        string
		metricName  string
		metric      repositories.Metric
		prepareMock func()
		wantErr     bool
	}{
		{
			name:       "update gauge metric",
			metricName: "gauge1",
			metric:     repositories.Metric{Type: "gauge", Value: float64(42.42)},
			prepareMock: func() {
				mock.ExpectExec(regexp.QuoteMeta(`
            INSERT INTO metrics (id, type, delta, value)
            VALUES ($1, 'gauge', NULL, $2)
            ON CONFLICT (id) DO UPDATE 
            SET delta = NULL,
                value = $2`)).
					WithArgs("gauge1", 42.42).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:       "update counter metric",
			metricName: "counter1",
			metric:     repositories.Metric{Type: "counter", Value: int64(10)},
			prepareMock: func() {
				mock.ExpectExec(regexp.QuoteMeta(`
            INSERT INTO metrics (id, type, delta, value)
            VALUES ($1, 'counter', $2, NULL)
            ON CONFLICT (id) DO UPDATE 
            SET delta = metrics.delta + $2,
                value = NULL`)).
					WithArgs("counter1", int64(10)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:        "unknown metric type error",
			metricName:  "unknown",
			metric:      repositories.Metric{Type: "unknown", Value: 0},
			prepareMock: func() {},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()
			err := storage.UpdateMetric(ctx, tt.metricName, tt.metric)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBStorage_GetMetric(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if errDB := db.Close(); errDB != nil {
			fmt.Printf("error closing db")
		}
	}()

	expectTableCreation(mock)
	storage, err := dbstorage.NewDBStorage(db)
	require.NoError(t, err)

	tests := []struct {
		name        string
		metricName  string
		prepareMock func()
		wantMetric  repositories.Metric
		wantErr     bool
	}{
		{
			name:       "get gauge metric",
			metricName: "gauge1",
			prepareMock: func() {
				rows := sqlmock.NewRows([]string{"type", "delta", "value"}).
					AddRow("gauge", nil, 123.456)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT type, delta, value FROM metrics WHERE id = $1`)).
					WithArgs("gauge1").
					WillReturnRows(rows)
			},
			wantMetric: repositories.Metric{Type: "gauge", Value: float64(123.456)},
		},
		{
			name:       "get counter metric",
			metricName: "counter1",
			prepareMock: func() {
				rows := sqlmock.NewRows([]string{"type", "delta", "value"}).
					AddRow("counter", int64(10), nil)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT type, delta, value FROM metrics WHERE id = $1`)).
					WithArgs("counter1").
					WillReturnRows(rows)
			},
			wantMetric: repositories.Metric{Type: "counter", Value: int64(10)},
		},
		{
			name:       "metric not found returns error",
			metricName: "missing",
			prepareMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT type, delta, value FROM metrics WHERE id = $1`)).
					WithArgs("missing").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()
			metric, err := storage.GetMetric(ctx, tt.metricName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMetric, metric)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBStorage_GetMetrics(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if errDB := db.Close(); errDB != nil {
			fmt.Printf("error closing db")
		}
	}()

	expectTableCreation(mock)
	storage, err := dbstorage.NewDBStorage(db)
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
		AddRow("g1", "gauge", nil, 10.5).
		AddRow("c1", "counter", int64(7), nil)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, type, delta, value FROM metrics`)).
		WillReturnRows(rows)

	got, err := storage.GetMetrics(ctx)
	require.NoError(t, err)

	want := map[string]repositories.Metric{
		"g1": {Type: "gauge", Value: float64(10.5)},
		"c1": {Type: "counter", Value: int64(7)},
	}

	assert.Equal(t, want, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBStorage_UpdateMetricsBatch(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if errDB := db.Close(); errDB != nil {
			fmt.Printf("error closing db")
		}
	}()

	expectTableCreation(mock)
	storage, err := dbstorage.NewDBStorage(db)
	require.NoError(t, err)

	mock.ExpectBegin()

	mock.ExpectPrepare(regexp.QuoteMeta(`
        INSERT INTO metrics (id, type, delta, value)
        VALUES ($1, 'counter', $2, NULL)
        ON CONFLICT (id) DO UPDATE 
        SET delta = metrics.delta + $2,
        value = NULL`))
	mock.ExpectPrepare(regexp.QuoteMeta(`
        INSERT INTO metrics (id, type, delta, value)
        VALUES ($1, 'gauge', NULL, $2)
        ON CONFLICT (id) DO UPDATE 
        SET delta = NULL,
            value = $2`))

	batch := map[string]repositories.Metric{
		"g1": {Type: "gauge", Value: float64(1.23)},
		"c1": {Type: "counter", Value: int64(5)},
	}

	mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metrics (id, type, delta, value)
        VALUES ($1, 'gauge', NULL, $2)
        ON CONFLICT (id) DO UPDATE 
        SET delta = NULL,
            value = $2`)).
		WithArgs("g1", float64(1.23)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metrics (id, type, delta, value)
        VALUES ($1, 'counter', $2, NULL)
        ON CONFLICT (id) DO UPDATE 
        SET delta = metrics.delta + $2,
        value = NULL`)).
		WithArgs("c1", int64(5)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = storage.UpdateMetricsBatch(ctx, batch)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
