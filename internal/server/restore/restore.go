package restore

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/a2sh3r/sysmetrics/internal/logger"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
	"go.uber.org/zap"
)

type RConfig struct {
	Interval int64
	FilePath string
	Storage  repositories.Storage
	mu       sync.Mutex
}

type metricData struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

var ErrRestoreFromFile = errors.New("error restoring from file")

func NewRestoreConfig(interval int64, filePath string, storage repositories.Storage) *RConfig {
	return &RConfig{
		Interval: interval,
		FilePath: filePath,
		Storage:  storage,
	}
}

func (b *RConfig) StartRestore(ctx context.Context) error {
	logger.Log.Info("Starting restore service",
		zap.Int64("interval_seconds", b.Interval),
		zap.String("file_path", b.FilePath))

	restoreTicker := time.NewTicker(time.Duration(b.Interval) * time.Second)
	defer restoreTicker.Stop()

	for {
		select {
		case <-restoreTicker.C:
			if err := b.SaveToFile(); err != nil {
				logger.Log.Error("Failed to save metrics to file", zap.Error(err))
			}
		case <-ctx.Done():
			if err := b.SaveToFile(); err != nil {
				return err
			}
			return nil
		}
	}
}

func (b *RConfig) SaveToFile() error {
	ctx := context.Background()
	b.mu.Lock()
	defer b.mu.Unlock()

	logger.Log.Debug("Saving metrics to file", zap.String("file", b.FilePath))

	dir := filepath.Dir(b.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	metrics, err := b.Storage.GetMetrics(ctx)
	if err != nil {
		return err
	}

	serializedMetrics := make(map[string]metricData)
	for name, metric := range metrics {
		serializedMetrics[name] = metricData{
			Type:  metric.Type,
			Value: metric.Value,
		}
	}

	file, err := os.OpenFile(b.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close file.Close: %v", err)
		}
	}()

	if err := json.NewEncoder(file).Encode(serializedMetrics); err != nil {
		logger.Log.Error("Error encoding metrics to JSON", zap.Error(err))
		return err
	}
	return nil
}

func RestoreFromFile(filename string) (*memstorage.MemStorage, error) {
	ctx := context.Background()
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logger.Log.Info("File does not exist, creating new storage", zap.String("filename", filename))
		return memstorage.NewMemStorage(), nil
	}

	file, err := os.Open(filename)
	if err != nil {
		logger.Log.Error("Error opening file", zap.String("filename", filename), zap.Error(err))
		return nil, ErrRestoreFromFile
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close file.Close: %v", err)
		}
	}()

	var serializedMetrics map[string]metricData
	if err := json.NewDecoder(file).Decode(&serializedMetrics); err != nil {
		logger.Log.Error("Error decoding metrics", zap.Error(err))
		return nil, ErrRestoreFromFile
	}

	ms := memstorage.NewMemStorage()

	for name, data := range serializedMetrics {
		var value interface{}
		switch data.Type {
		case constants.MetricTypeCounter:
			if floatVal, ok := data.Value.(float64); ok {
				value = int64(floatVal)
			} else {
				logger.Log.Warn("Invalid counter value type", zap.String("name", name), zap.Any("value", data.Value))
				continue
			}
		case constants.MetricTypeGauge:
			if floatVal, ok := data.Value.(float64); ok {
				value = floatVal
			} else {
				logger.Log.Warn("Invalid gauge value type", zap.String("name", name), zap.Any("value", data.Value))
				continue
			}
		default:
			logger.Log.Warn("Unknown metric type", zap.String("name", name), zap.String("type", data.Type))
			continue
		}

		err := ms.UpdateMetric(ctx, name, repositories.Metric{
			Type:  data.Type,
			Value: value,
		})
		if err != nil {
			logger.Log.Warn("Failed to restore metric", zap.String("name", name), zap.Error(err))
		}
	}

	return ms, nil
}
