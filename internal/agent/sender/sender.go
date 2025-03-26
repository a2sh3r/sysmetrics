package sender

import (
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"io"
	"log"
	"net/http"
	"reflect"
)

type Sender struct {
	serverAddress string
}

func NewSender(serverAddress string) *Sender {
	return &Sender{
		serverAddress: serverAddress,
	}
}

func (s *Sender) sendMetric(metricType, metricName string, value interface{}) error {
	var strValue string
	switch v := value.(type) {
	case int64:
		strValue = fmt.Sprintf("%d", v)
	case float64:
		strValue = fmt.Sprintf("%f", v)
	default:
		return fmt.Errorf("unsupported metric type: %T", v)
	}

	url := fmt.Sprintf("%s/update/%s/%s/%s", s.serverAddress, metricType, metricName, strValue)

	res, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("error closing response body: ", err)
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("Server response (status %d): %s", res.StatusCode, string(body))

	return nil
}

func (s *Sender) SendMetrics(metricsBatch []*metrics.Metrics) error {
	for _, m := range metricsBatch {
		v := reflect.ValueOf(m).Elem()
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			fieldName := t.Field(i).Name

			var metricType string
			if fieldName == "PollCount" {
				metricType = "counter"
			} else {
				metricType = "gauge"
			}

			var value interface{}
			switch fieldValue.Kind() {
			case reflect.Int64:
				value = fieldValue.Int()
			case reflect.Float64:
				value = fieldValue.Float()
			default:
				return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
			}

			err := s.sendMetric(metricType, fieldName, value)
			if err != nil {
				return fmt.Errorf("failed to send metric %s: %w", fieldName, err)
			}
		}
	}
	return nil
}
