package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
		want *AgentConfig
	}{
		{
			name: "Test #1 create valid config",
			want: &AgentConfig{
				PollInterval:   2,
				ReportInterval: 10,
				Address:        "http://localhost:8080",
				RateLimit:      1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAgentConfig()
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewServerConfig(t *testing.T) {
	tests := []struct {
		name     string
		setEnv   map[string]string
		unsetEnv []string
		wantErr  bool
	}{
		{"valid env", map[string]string{"ADDRESS": "localhost:9999"}, nil, false},
		{"invalid env", map[string]string{"STORE_INTERVAL": "notanint"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.setEnv {
				_ = os.Setenv(k, v)
			}
			for _, k := range tt.unsetEnv {
				_ = os.Unsetenv(k)
			}
			cfg, err := NewServerConfig()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
			}
			for k := range tt.setEnv {
				_ = os.Unsetenv(k)
			}
		})
	}
}
