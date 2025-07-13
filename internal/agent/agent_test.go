package agent

import (
	"context"
	"testing"
	"time"

	"github.com/a2sh3r/sysmetrics/internal/agent/collector"
	"github.com/a2sh3r/sysmetrics/internal/agent/sender"
	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestAgent_Run(t *testing.T) {
	type fields struct {
		collector      *collector.Collector
		sender         *sender.Sender
		pollInterval   time.Duration
		reportInterval time.Duration
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
			name: "Test #1 run agent with valid intervals",
			fields: fields{
				collector:      &collector.Collector{},
				sender:         sender.NewSender("http://localhost:8080"),
				pollInterval:   time.Second,
				reportInterval: time.Second * 10,
			},
			args: args{
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				collector:      tt.fields.collector,
				sender:         tt.fields.sender,
				pollInterval:   tt.fields.pollInterval,
				reportInterval: tt.fields.reportInterval,
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
				},
			},
			want: &Agent{
				collector:      &collector.Collector{},
				sender:         sender.NewSender("http://localhost:8080"),
				pollInterval:   time.Duration(2) * time.Second,
				reportInterval: time.Duration(10) * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAgent(tt.args.cfg)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want, got)
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
