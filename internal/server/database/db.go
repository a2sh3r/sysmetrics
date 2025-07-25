// Package database provides functions for initializing and closing the database connection.
package database

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"

	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// InitDB initializes a new database connection using the provided server configuration.
func InitDB(cfg *config.ServerConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	logger.Log.Info("Successfully connected to the database", zap.Any("database dsn", cfg.DatabaseDSN))
	return db, nil
}

// CloseDB closes the provided database connection and logs the result.
func CloseDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		logger.Log.Error("Database connection closed", zap.Error(err))
	} else {
		logger.Log.Info("Database connection closed")
	}
}
