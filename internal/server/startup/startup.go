// Package startup provides the main server startup logic.
package startup

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"github.com/a2sh3r/sysmetrics/internal/server/database"
	"github.com/a2sh3r/sysmetrics/internal/server/handlers"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/restore"
	"github.com/a2sh3r/sysmetrics/internal/server/services"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/dbstorage"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
)

// RunServer starts the HTTP server with the provided configuration.
func RunServer(cfg *config.ServerConfig) error {

	var storage repositories.Storage
	var err error
	var db *sql.DB

	if cfg.DatabaseDSN != "" {
		db, err = database.InitDB(cfg)
		if err != nil {
			logger.Log.Error("Database connection failed", zap.Error(err))
			return err
		}
		defer database.CloseDB(db)
		storage, err = dbstorage.NewDBStorage(db)
		if err != nil {
			logger.Log.Error("Failed to initialize DBStorage", zap.Error(err))
			return err
		}
	} else {
		var memStorage *memstorage.MemStorage
		if cfg.Restore {
			memStorage, err = restore.RestoreFromFile(cfg.FileStoragePath)
			if err != nil {
				logger.Log.Info("Failed to restore metrics from file, using empty storage",
					zap.Error(err),
					zap.String("file", cfg.FileStoragePath))
				memStorage = memstorage.NewMemStorage()
			}
		} else {
			memStorage = memstorage.NewMemStorage()
		}
		storage = memStorage
	}

	metricRepo := repositories.NewMetricRepo(storage)
	metricService := services.NewService(metricRepo)
	handler := handlers.NewHandler(metricService, metricService, db)

	restoreConfig := restore.NewRestoreConfig(int64(cfg.StoreInterval), cfg.FileStoragePath, storage)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.StoreInterval != 0 {
		go func() {
			if err := restoreConfig.StartRestore(ctx); err != nil {
				logger.Log.Error("Restore service failed", zap.Error(err))
			}
		}()
	}

	srvMux := http.NewServeMux()
	srvMux.Handle("/debug/pprof/", http.DefaultServeMux)
	srvMux.Handle("/", handlers.NewRouter(handler, cfg))

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: srvMux,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-quit
		logger.Log.Info("Received shutdown signal", zap.String("signal", sig.String()))
		logger.Log.Info("Shutting down server gracefully...")

		cancel()

		if err := restoreConfig.SaveToFile(); err != nil {
			logger.Log.Error("Error saving metrics on shutdown", zap.Error(err))
		} else {
			logger.Log.Info("Metrics successfully saved before shutdown")
		}

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Log.Error("Server shutdown error", zap.Error(err))
		} else {
			logger.Log.Info("HTTP server gracefully stopped")
		}
	}()

	logger.Log.Info("Server is starting",
		zap.String("address", cfg.Address),
		zap.Int("store_interval", cfg.StoreInterval),
		zap.String("storage_path", cfg.FileStoragePath),
		zap.Bool("restore", cfg.Restore))

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-ctx.Done()
	logger.Log.Info("Server shutdown completed")

	return nil
}
