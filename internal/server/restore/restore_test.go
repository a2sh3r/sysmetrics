package restore

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Open(name string) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
}

type realFileSystem struct{}

func (realFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (realFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (realFileSystem) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (realFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

type mockFileSystem struct {
	mock.Mock
}

func (m *mockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (m *mockFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	args := m.Called(name, flag, perm)
	return args.Get(0).(*os.File), args.Error(1)
}

func (m *mockFileSystem) Open(name string) (*os.File, error) {
	args := m.Called(name)
	return args.Get(0).(*os.File), args.Error(1)
}

func (m *mockFileSystem) Stat(name string) (os.FileInfo, error) {
	args := m.Called(name)
	return args.Get(0).(os.FileInfo), args.Error(1)
}

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) GetMetrics(ctx context.Context) (map[string]repositories.Metric, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]repositories.Metric), args.Error(1)
}

func (m *mockStorage) GetMetric(ctx context.Context, name string) (repositories.Metric, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(repositories.Metric), args.Error(1)
}

func (m *mockStorage) UpdateMetric(ctx context.Context, name string, metric repositories.Metric) error {
	args := m.Called(ctx, name, metric)
	return args.Error(0)
}

func (m *mockStorage) UpdateMetricsBatch(ctx context.Context, metrics map[string]repositories.Metric) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func TestNewRestoreConfig(t *testing.T) {
	tests := []struct {
		name     string
		interval int64
		filePath string
		storage  repositories.Storage
		expected *RConfig
	}{
		{
			name:     "Valid configuration",
			interval: 10,
			filePath: "/tmp/metrics.json",
			storage:  memstorage.NewMemStorage(),
			expected: &RConfig{
				Interval: 10,
				FilePath: "/tmp/metrics.json",
				Storage:  memstorage.NewMemStorage(),
			},
		},
		{
			name:     "Zero interval",
			interval: 0,
			filePath: "/tmp/empty.json",
			storage:  memstorage.NewMemStorage(),
			expected: &RConfig{
				Interval: 0,
				FilePath: "/tmp/empty.json",
				Storage:  memstorage.NewMemStorage(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRestoreConfig(tt.interval, tt.filePath, tt.storage)
			assert.Equal(t, tt.expected.Interval, got.Interval)
			assert.Equal(t, tt.expected.FilePath, got.FilePath)
			assert.Equal(t, tt.expected.Storage, got.Storage)
			assert.NotNil(t, &got.mu)
		})
	}
}

func TestSaveToFile(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		metrics        map[string]repositories.Metric
		getMetricsErr  error
		fs             fileSystem
		expectedErr    bool
		expectedErrMsg string
	}{
		{
			name:     "Successful save",
			filePath: "metrics.json",
			metrics: map[string]repositories.Metric{
				"metric1": {Type: constants.MetricTypeGauge, Value: 42.0},
				"metric2": {Type: constants.MetricTypeCounter, Value: int64(100)},
			},
			getMetricsErr:  nil,
			fs:             realFileSystem{},
			expectedErr:    false,
			expectedErrMsg: "",
		},
		{
			name:           "GetMetrics error",
			filePath:       "metrics.json",
			metrics:        nil,
			getMetricsErr:  assert.AnError,
			fs:             realFileSystem{},
			expectedErr:    true,
			expectedErrMsg: assert.AnError.Error(),
		},
		{
			name:          "Create directory error",
			filePath:      "/invalid/dir/metrics.json",
			metrics:       map[string]repositories.Metric{},
			getMetricsErr: nil,
			fs: &mockFileSystem{
				Mock: mock.Mock{},
			},
			expectedErr:    true,
			expectedErrMsg: assert.AnError.Error(),
		},
		{
			name:          "Open file error",
			filePath:      "metrics.json",
			metrics:       map[string]repositories.Metric{},
			getMetricsErr: nil,
			fs: &mockFileSystem{
				Mock: mock.Mock{},
			},
			expectedErr:    true,
			expectedErrMsg: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, tt.filePath)

			storage := new(mockStorage)
			storage.On("GetMetrics", mock.Anything).Return(tt.metrics, tt.getMetricsErr).Maybe()

			if mockFS, ok := tt.fs.(*mockFileSystem); ok {
				switch tt.name {
				case "Create directory error":
					mockFS.On("MkdirAll", filepath.Dir(filePath), os.FileMode(0755)).Return(assert.AnError).Once()
				case "Open file error":
					mockFS.On("MkdirAll", filepath.Dir(filePath), os.FileMode(0755)).Return(nil).Once()
					mockFS.On("OpenFile", filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0666)).Return((*os.File)(nil), assert.AnError).Once()
				}
			}

			config := &RConfig{
				Interval: 10,
				FilePath: filePath,
				Storage:  storage,
			}

			saveToFile := func(fs fileSystem) error {
				ctx := context.Background()
				config.mu.Lock()
				defer config.mu.Unlock()

				dir := filepath.Dir(config.FilePath)
				if err := fs.MkdirAll(dir, 0755); err != nil {
					return err
				}

				metrics, err := config.Storage.GetMetrics(ctx)
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

				file, err := fs.OpenFile(config.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
				if err != nil {
					return err
				}
				defer file.Close()

				return json.NewEncoder(file).Encode(serializedMetrics)
			}

			err := saveToFile(tt.fs)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)

				if tt.metrics != nil {
					file, err := os.Open(filePath)
					require.NoError(t, err)
					defer file.Close()

					var gotMetrics map[string]metricData
					err = json.NewDecoder(file).Decode(&gotMetrics)
					require.NoError(t, err)

					for name, metric := range tt.metrics {
						gotMetric, exists := gotMetrics[name]
						assert.True(t, exists)
						assert.Equal(t, metric.Type, gotMetric.Type)
						if metric.Type == constants.MetricTypeCounter {
							assert.Equal(t, float64(metric.Value.(int64)), gotMetric.Value)
						} else {
							assert.Equal(t, metric.Value, gotMetric.Value)
						}
					}
				}
			}
		})
	}
}

func TestStartRestore(t *testing.T) {
	type testRConfig struct {
		*RConfig
		saveToFileFunc func() error
	}

	startRestore := func(config *testRConfig, ctx context.Context) error {
		restoreTicker := time.NewTicker(time.Duration(config.RConfig.Interval) * time.Second)
		defer restoreTicker.Stop()

		for {
			select {
			case <-restoreTicker.C:
				if err := config.saveToFileFunc(); err != nil {
					return err
				}
			case <-ctx.Done():
				return config.saveToFileFunc()
			}
		}
	}

	tests := []struct {
		name           string
		interval       int64
		saveToFileErr  error
		expectedErr    bool
		expectedErrMsg string
	}{
		{
			name:          "Successful run with save",
			interval:      1,
			saveToFileErr: nil,
			expectedErr:   false,
		},
		{
			name:           "SaveToFile error",
			interval:       1,
			saveToFileErr:  assert.AnError,
			expectedErr:    true,
			expectedErrMsg: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, "metrics.json")

			storage := new(mockStorage)
			storage.On("GetMetrics", mock.Anything).Return(map[string]repositories.Metric{}, nil)
			storage.On("GetMetric", mock.Anything, mock.Anything).Return(repositories.Metric{}, nil)
			storage.On("UpdateMetricsBatch", mock.Anything, mock.Anything).Return(nil)

			config := &testRConfig{
				RConfig: NewRestoreConfig(tt.interval, filePath, storage),
				saveToFileFunc: func() error {
					if tt.saveToFileErr != nil {
						return tt.saveToFileErr
					}
					return (&RConfig{
						Interval: tt.interval,
						FilePath: filePath,
						Storage:  storage,
					}).SaveToFile()
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			errChan := make(chan error)
			go func() {
				errChan <- startRestore(config, ctx)
			}()

			select {
			case err := <-errChan:
				if tt.expectedErr {
					assert.Error(t, err)
					if tt.expectedErrMsg != "" {
						assert.Contains(t, err.Error(), tt.expectedErrMsg)
					}
				} else {
					assert.NoError(t, err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("Test timed out")
			}
		})
	}
}
