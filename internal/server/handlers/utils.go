package handlers

import (
	"errors"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/models"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"net/http"
	"strings"
	"time"
)

func validateParams(params ...string) error {
	for _, p := range params {
		if strings.TrimSpace(p) == "" {
			return errors.New("one of metric parameters are null")
		}
	}
	return nil
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Date", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
}

func formatMetric(metricName *string, value interface{}) (string, error) {
	switch v := value.(type) {
	case int64:
		if metricName != nil {
			return fmt.Sprintf("%s %d\n", *metricName, v), nil
		}
		return fmt.Sprintf("%d\n", v), nil
	case float64:
		if metricName != nil {
			return fmt.Sprintf("%s %g\n", *metricName, v), nil
		}
		return fmt.Sprintf("%g\n", v), nil
	default:
		return "", errors.New("unsupported metric value type")
	}
}

func floatPointer(f float64) *float64 {
	return &f
}

func convertMetricToModel(id string, metric interface{}) models.Metrics {
	switch m := metric.(type) {
	case repositories.Metric:
		result := models.Metrics{
			ID:    id,
			MType: m.Type,
		}
		switch m.Type {
		case constants.MetricTypeGauge:
			if v, ok := m.Value.(float64); ok {
				result.Value = &v
			}
		case constants.MetricTypeCounter:
			if v, ok := m.Value.(int64); ok {
				result.Delta = &v
			}
		}
		return result
	}
	return models.Metrics{}
}
