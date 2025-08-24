// Package agent implements the main agent logic for collecting and sending metrics.
package agent

import (
	"context"
	"sync"
	"time"

	grpcclient "github.com/a2sh3r/sysmetrics/internal/agent/grpc"
	"github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/a2sh3r/sysmetrics/internal/agent/sender"
	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	"go.uber.org/zap"
)

// Agent represents the metrics agent.
type Agent struct {
	cfg        *config.AgentConfig
	metrics    *metrics.Metrics
	worker     *MetricsWorker
	sender     *sender.Sender
	grpcClient *grpcclient.Client
	mu         sync.RWMutex
}

// NewAgent creates a new Agent instance.
func NewAgent(cfg *config.AgentConfig) *Agent {
	agent := &Agent{
		cfg:     cfg,
		metrics: metrics.NewMetrics(),
	}

	logger.Log.Info("Initializing agent",
		zap.String("protocol", cfg.Protocol),
		zap.String("address", cfg.Address),
		zap.String("grpc_address", cfg.GRPCAddress),
		zap.Duration("poll_interval", cfg.PollInterval.Duration),
		zap.Duration("report_interval", cfg.ReportInterval.Duration),
		zap.Int64("rate_limit", cfg.RateLimit))

	if cfg.Protocol == "grpc" {
		grpcClient, err := grpcclient.NewClient(cfg.GRPCAddress, cfg.SecretKey, cfg.CryptoKey)
		if err != nil {
			logger.Log.Error("Failed to create gRPC client", zap.Error(err))
		} else {
			agent.grpcClient = grpcClient
			logger.Log.Info("gRPC client initialized successfully")
		}
	} else {
		agent.sender = sender.NewSender(cfg.Address, cfg.SecretKey, cfg.CryptoKey)
		logger.Log.Info("HTTP sender initialized successfully")
	}

	return agent
}

// Run starts the agent's main loop.
func (a *Agent) Run(ctx context.Context) {
	logger.Log.Info("Starting agent main loop")
	
	a.worker = NewMetricsWorker(a.cfg.RateLimit, a.sendMetrics)
	a.worker.Start(ctx)

	metricsTicker := time.NewTicker(a.cfg.PollInterval.Duration)
	defer metricsTicker.Stop()

	systemTicker := time.NewTicker(a.cfg.PollInterval.Duration)
	defer systemTicker.Stop()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("Metrics collection stopped")
				return
			case <-metricsTicker.C:
				a.mu.Lock()
				a.metrics = metrics.NewMetrics()
				a.mu.Unlock()
				a.worker.SendMetrics(a.metrics)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("System metrics update stopped")
				return
			case <-systemTicker.C:
				a.mu.Lock()
				if err := a.metrics.UpdateSystemMetrics(); err != nil {
					logger.Log.Error("Error updating system metrics", zap.Error(err))
				}
				a.mu.Unlock()
				a.worker.SendMetrics(a.metrics)
			}
		}
	}()

	<-ctx.Done()
	logger.Log.Info("Agent shutdown initiated, waiting for operations to complete...")

	a.worker.Stop()

	<-done
	logger.Log.Info("All agent operations completed")
}

// sendMetrics sends collected metrics to the server.
func (a *Agent) sendMetrics(m *metrics.Metrics) error {
	logger.Log.Debug("Sending metrics to server",
		zap.String("protocol", a.cfg.Protocol),
		zap.Float64("alloc", m.Alloc),
		zap.Float64("heap_alloc", m.HeapAlloc),
		zap.Int64("poll_count", m.PollCount))

	if a.cfg.Protocol == "grpc" && a.grpcClient != nil {
		err := a.grpcClient.SendMetricsBatch(context.Background(), m)
		if err != nil {
			logger.Log.Error("Failed to send metrics via gRPC", zap.Error(err))
		} else {
			logger.Log.Debug("Metrics sent successfully via gRPC")
		}
		return err
	}

	err := a.sender.SendMetricsWithRetries(context.Background(), []*metrics.Metrics{m})
	if err != nil {
		logger.Log.Error("Failed to send metrics via HTTP", zap.Error(err))
	} else {
		logger.Log.Debug("Metrics sent successfully via HTTP")
	}
	return err
}
