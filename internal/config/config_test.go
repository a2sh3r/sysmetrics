package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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
