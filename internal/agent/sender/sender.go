package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/a2sh3r/sysmetrics/internal/agent/utils"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/hash"
	"github.com/a2sh3r/sysmetrics/internal/models"
	"github.com/a2sh3r/sysmetrics/internal/server/middleware"
	"io"
	"log"
	"net/http"
	"reflect"
	"time"
)

type Sender struct {
	serverAddress string
	client        *http.Client
	secretKey     string
}

func NewSender(serverAddress string, secretKey string) *Sender {
	return &Sender{
		serverAddress: serverAddress,
		client:        &http.Client{},
		secretKey:     secretKey,
	}
}

func toModelMetrics(m *metrics.Metrics) []*models.Metrics {
	var result []*models.Metrics
	val := reflect.ValueOf(m).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name

		switch field.Kind() {
		case reflect.Float64:
			fv := field.Float()
			result = append(result, &models.Metrics{
				ID:    fieldName,
				MType: constants.MetricTypeGauge,
				Delta: nil,
				Value: &fv,
			})
		case reflect.Int64:
			iv := field.Int()
			result = append(result, &models.Metrics{
				ID:    fieldName,
				MType: constants.MetricTypeCounter,
				Delta: &iv,
				Value: nil,
			})
		case reflect.Slice:
			if fieldName == "CPUUtilization" {
				for j := 0; j < field.Len(); j++ {
					fv := field.Index(j).Float()
					result = append(result, &models.Metrics{
						ID:    fmt.Sprintf("CPUutilization%d", j+1),
						MType: constants.MetricTypeGauge,
						Delta: nil,
						Value: &fv,
					})
				}
			}
		default:
			panic("unhandled default case")
		}
	}

	return result
}

func (s *Sender) sendMetricsBatchJSON(ctx context.Context, metrics []*models.Metrics) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics batch: %w", err)
	}

	compressedData, err := utils.CompressData(data)
	if err != nil {
		return fmt.Errorf("failed to compress metrics batch: %w", err)
	}

	url := s.serverAddress + "/updates/"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(compressedData))
	if err != nil {
		return fmt.Errorf("failed to create batch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	if s.secretKey != "" {
		calculateHash := hash.CalculateHash(string(data), s.secretKey)
		req.Header.Set(middleware.HashHeader, calculateHash)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send batch request: %w", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				log.Printf("error closing response body: %v", err)
			}
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("Server batch response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d for batch update", resp.StatusCode)
	}

	return nil
}

func (s *Sender) SendMetrics(ctx context.Context, metricsBatch []*metrics.Metrics) error {
	if metricsBatch == nil {
		return fmt.Errorf("metricsBatch is nil")
	}
	if len(metricsBatch) == 0 {
		return fmt.Errorf("metricsBatch is empty")
	}

	var allModelMetrics []*models.Metrics
	for _, m := range metricsBatch {
		if m == nil {
			return fmt.Errorf("metric is nil")
		}
		modelMetrics := toModelMetrics(m)
		allModelMetrics = append(allModelMetrics, modelMetrics...)
	}

	return s.sendMetricsBatchJSON(ctx, allModelMetrics)
}

func (s *Sender) SendMetricsWithRetries(ctx context.Context, metricsBatch []*metrics.Metrics) error {
	retries := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	var lastErr error
	for _, wait := range retries {
		if err := s.SendMetrics(ctx, metricsBatch); err != nil {
			log.Printf("retriable error: %v, retrying in %v", err, wait)
			lastErr = err

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
				continue
			}

		} else {
			return nil
		}
	}

	return lastErr
}
