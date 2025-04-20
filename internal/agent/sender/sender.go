package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/a2sh3r/sysmetrics/internal/agent/utils"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/models"
	"io"
	"log"
	"net/http"
	"reflect"
)

type Sender struct {
	serverAddress string
	client        *http.Client
}

func NewSender(serverAddress string) *Sender {
	return &Sender{
		serverAddress: serverAddress,
		client:        &http.Client{},
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
		default:
			panic("unhandled default case")
		}
	}

	return result
}

func (s *Sender) sendMetricJSON(metric *models.Metrics) error {
	data, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric %s: %w", metric.ID, err)
	}

	compressedData, err := utils.CompressData(data)
	if err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}

	url := s.serverAddress + "/update/"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(compressedData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
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

	log.Printf("Server response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d for metric %s", resp.StatusCode, metric.ID)
	}

	return nil
}

func (s *Sender) SendMetrics(metricsBatch []*metrics.Metrics) error {
	if metricsBatch == nil {
		return fmt.Errorf("metricsBatch is nil")
	}

	if len(metricsBatch) == 0 {
		return fmt.Errorf("metricsBatch is empty")
	}

	for _, m := range metricsBatch {
		if m == nil {
			return fmt.Errorf("metric is nil")
		}
		modelMetrics := toModelMetrics(m)
		for _, metric := range modelMetrics {
			err := s.sendMetricJSON(metric)
			if err != nil {
				return fmt.Errorf("failed to send metric %s: %w", metric.ID, err)
			}
		}
	}

	return nil
}
