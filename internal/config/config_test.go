package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
		want *AgentConfig
	}{
		{
			name: "Test #1 create valid config",
			want: &AgentConfig{
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
				Address:        "http://localhost:8080",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAgentConfig()
			assert.NotNil(t, got)
			assert.Equal(t, tt.want, got)
		})
	}
}
