package startup

import (
	"context"
	"errors"
	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"github.com/a2sh3r/sysmetrics/internal/server/handlers"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/restore"
	"github.com/a2sh3r/sysmetrics/internal/server/services"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunServer(cfg *config.ServerConfig) error {

	var memStorage *memstorage.MemStorage
	var err error

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

	metricRepo := repositories.NewMetricRepo(memStorage)
	metricService := services.NewService(metricRepo)
	handler := handlers.NewHandler(metricService, metricService)

	restoreConfig := restore.NewRestoreConfig(int64(cfg.StoreInterval), cfg.FileStoragePath, memStorage)

	if cfg.StoreInterval != 0 {
		go func() {
			if err := restoreConfig.StartRestore(context.Background()); err != nil {
				logger.Log.Error("Restore service failed", zap.Error(err))
			}
		}()
	}

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: handlers.NewRouter(handler),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Log.Info("Shutting down server...")

		if err := restoreConfig.SaveToFile(); err != nil {
			logger.Log.Error("Error saving metrics on shutdown", zap.Error(err))
		} else {
			logger.Log.Info("Metrics successfully saved before shutdown")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Log.Error("Server shutdown error", zap.Error(err))
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

	return nil
}
