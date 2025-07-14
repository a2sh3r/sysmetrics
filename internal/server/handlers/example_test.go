package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/server/handlers"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
)

type mockService struct {
	metrics map[string]repositories.Metric
}

func (m *mockService) GetMetricWithRetry(_ context.Context, name string) (repositories.Metric, error) {
	metric, ok := m.metrics[name]
	if !ok {
		return repositories.Metric{}, http.ErrNoLocation
	}
	return metric, nil
}
func (m *mockService) GetMetricsWithRetry(_ context.Context) (map[string]repositories.Metric, error) {
	return m.metrics, nil
}
func (m *mockService) UpdateGaugeMetricWithRetry(_ context.Context, name string, value float64) error {
	m.metrics[name] = repositories.Metric{Type: constants.MetricTypeGauge, Value: value}
	return nil
}
func (m *mockService) UpdateCounterMetricWithRetry(_ context.Context, name string, value int64) error {
	m.metrics[name] = repositories.Metric{Type: constants.MetricTypeCounter, Value: value}
	return nil
}
func (m *mockService) UpdateMetricsBatchWithRetry(_ context.Context, metrics map[string]repositories.Metric) error {
	for k, v := range metrics {
		m.metrics[k] = v
	}
	return nil
}

func newTestServer() (*mockService, *httptest.Server) {
	svc := &mockService{metrics: make(map[string]repositories.Metric)}
	h := handlers.NewHandler(svc, svc, nil)
	r := handlers.NewRouter(h, &config.ServerConfig{})
	ts := httptest.NewServer(r)
	return svc, ts
}

func ExampleHandler_UpdateMetric_gauge() {
	_, ts := newTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/update/gauge/testGauge/123.45", nil)
	resp, _ := http.DefaultClient.Do(req)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close resp.Body: %v", err)
		}
	}()

	fmt.Println(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	// Output:
	// 200
	// Metric testGauge is updated successfully with value 123.45
}

func ExampleHandler_UpdateMetric_counter() {
	_, ts := newTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/update/counter/testCounter/42", nil)
	resp, _ := http.DefaultClient.Do(req)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close resp.Body: %v", err)
		}
	}()

	fmt.Println(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	// Output:
	// 200
	// Metric testCounter is updated successfully with value 42
}

func ExampleHandler_UpdateSerializedMetric() {
	_, ts := newTestServer()
	defer ts.Close()

	body := bytes.NewBufferString(`{"id":"Alloc","type":"gauge","value":123.45}`)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/update/", body)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close resp.Body: %v", err)
		}
	}()

	fmt.Println(resp.StatusCode)
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":123.45}
}

func ExampleHandler_GetMetric() {
	svc, ts := newTestServer()
	defer ts.Close()

	// Сначала добавим метрику через update
	err := svc.UpdateGaugeMetricWithRetry(context.Background(), "testGauge", 123.45)
	if err != nil {
		return
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/value/gauge/testGauge", nil)
	resp, _ := http.DefaultClient.Do(req)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close resp.Body: %v", err)
		}
	}()

	fmt.Println(resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	// Output:
	// 200
	// 123.45
}

func ExampleHandler_GetSerializedMetric() {
	svc, ts := newTestServer()
	defer ts.Close()

	// Сначала добавим метрику через update
	err := svc.UpdateGaugeMetricWithRetry(context.Background(), "Alloc", 123.45)
	if err != nil {
		return
	}

	body := bytes.NewBufferString(`{"id":"Alloc","type":"gauge"}`)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/value/", body)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close resp.Body: %v", err)
		}
	}()

	fmt.Println(resp.StatusCode)
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":123.45}
}

func ExampleHandler_UpdateSerializedMetrics() {
	_, ts := newTestServer()
	defer ts.Close()

	metrics := []map[string]interface{}{
		{"id": "Alloc", "type": "gauge", "value": 123.45},
		{"id": "Count", "type": "counter", "delta": 10},
	}
	body, _ := json.Marshal(metrics)
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/updates/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close resp.Body: %v", err)
		}
	}()

	fmt.Println(resp.StatusCode)
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	// Output:
	// 200
	//
}

func ExampleHandler_Ping() {
	_, ts := newTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/ping", nil)
	resp, _ := http.DefaultClient.Do(req)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close resp.Body: %v", err)
		}
	}()

	fmt.Println(resp.StatusCode)
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	// Output:
	// 200
	//
}
