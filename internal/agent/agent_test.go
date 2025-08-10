package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/a2sh3r/sysmetrics/internal/agent/sender"
	"github.com/a2sh3r/sysmetrics/internal/config"
)

func TestAgent_Run(t *testing.T) {
	type fields struct {
		cfg     *config.AgentConfig
		sender  *sender.Sender
		metrics *metrics.Metrics
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Test #1 run agent with valid config",
			fields: fields{
				cfg: &config.AgentConfig{
					Address:        "http://localhost:8080",
					PollInterval:   2,
					ReportInterval: 10,
					SecretKey:      "test key",
					RateLimit:      1,
				},
				metrics: metrics.NewMetrics(),
				sender:  sender.NewSender("http://localhost:8080", "test key", "test key"),
			},
			args: args{
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				cfg:     tt.fields.cfg,
				metrics: tt.fields.metrics,
				sender:  tt.fields.sender,
			}
			ctx, cancel := context.WithCancel(tt.args.ctx)
			go func() {
				time.Sleep(2 * time.Millisecond)
				cancel()
			}()

			go a.Run(ctx)

			<-ctx.Done()

			assert.True(t, ctx.Err() != nil, "context was not canceled")
		})
	}
}

func TestNewAgent(t *testing.T) {
	type args struct {
		cfg *config.AgentConfig
	}
	tests := []struct {
		name string
		args args
		want *Agent
	}{
		{
			name: "Test #1 create agent with valid config",
			args: args{
				cfg: &config.AgentConfig{
					Address:        "http://localhost:8080",
					PollInterval:   2,
					ReportInterval: 10,
					SecretKey:      "test key",
					RateLimit:      1,
				},
			},
			want: &Agent{
				cfg: &config.AgentConfig{
					Address:        "http://localhost:8080",
					PollInterval:   2,
					ReportInterval: 10,
					SecretKey:      "test key",
					RateLimit:      1,
				},
				metrics: metrics.NewMetrics(),
				sender:  sender.NewSender("http://localhost:8080", "test key", ""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAgent(tt.args.cfg)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want.cfg, got.cfg)
			assert.NotNil(t, got.metrics)
			assert.NotNil(t, got.sender)
		})
	}
}

func BenchmarkAgentRun(b *testing.B) {
	cfg := &config.AgentConfig{
		Address:        "http://localhost:8080",
		PollInterval:   2,
		ReportInterval: 10,
	}
	agent := NewAgent(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go agent.Run(ctx)
		cancel()
	}
}
