package handlers

import (
	"github.com/a2sh3r/sysmetrics/internal/constants"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_GetMetrics(t *testing.T) {
	type fields struct {
		service *services.Service
	}
	type args struct {
		method string
		url    string
	}
	type want struct {
		code        int
		response    []string
		contentType string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "Test #1 get gauge metric",
			fields: fields{
				service: services.NewService(&mockRepo{
					metrics: map[string]repositories.Metric{
						"test_gauge": {
							Type:  constants.MetricTypeGauge,
							Value: 123.123,
						},
						"test_counter": {
							Type:  constants.MetricTypeCounter,
							Value: int64(123),
						},
					},
				}),
			},
			args: args{
				method: http.MethodGet,
				url:    "/",
			},
			want: want{
				code: http.StatusOK,
				response: []string{
					"test_gauge 123.123",
					"test_counter 123",
				},
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test #2 get invalid metric",
			fields: fields{
				service: services.NewService(&mockRepo{
					metrics: map[string]repositories.Metric{
						"test_invalid": {
							Type:  constants.MetricTypeGauge,
							Value: "String",
						},
					},
				}),
			},
			args: args{
				method: http.MethodGet,
				url:    "/",
			},
			want: want{
				code: http.StatusInternalServerError,
				response: []string{
					"unsupported metric value type",
				},
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
			}

			ts := httptest.NewServer(NewRouter(h))
			defer ts.Close()

			req, err := http.NewRequest(tt.args.method, ts.URL+tt.args.url, nil)
			require.NoError(t, err)

			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, tt.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			actualLines := strings.Split(strings.TrimSpace(string(resBody)), "\n")

			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			for _, expectedLine := range tt.want.response {
				assert.Contains(t, actualLines, expectedLine)
			}
			assert.Equal(t, len(tt.want.response), len(actualLines))
		})
	}
}
