// Package grpc provides gRPC client functionality for sending metrics to the server.
package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	agentmetrics "github.com/a2sh3r/sysmetrics/internal/agent/metrics"
	"github.com/a2sh3r/sysmetrics/internal/crypto"
	"github.com/a2sh3r/sysmetrics/internal/hash"
	"github.com/a2sh3r/sysmetrics/internal/logger"
	pb "github.com/a2sh3r/sysmetrics/pkg/api/grpc/metrics"
	"go.uber.org/zap"
)

// Client represents the gRPC client
type Client struct {
	conn      *grpc.ClientConn
	client    pb.MetricsServiceClient
	secretKey string
	encryptor *crypto.Encryptor
}

// NewClient creates a new gRPC client
func NewClient(addr string, secretKey string, cryptoKeyPath string) (*Client, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}

	var encryptor *crypto.Encryptor
	if cryptoKeyPath != "" {
		encryptor, err = crypto.NewEncryptor(cryptoKeyPath)
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		conn:      conn,
		client:    pb.NewMetricsServiceClient(conn),
		secretKey: secretKey,
		encryptor: encryptor,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// SendMetricsBatch sends a batch of metrics with metadata
func (c *Client) SendMetricsBatch(ctx context.Context, metrics interface{}) error {
	logger.Log.Info("Sending metrics batch via gRPC")

	pbMetrics := convertToProtoMetrics(metrics)

	logger.Log.Debug("Converted metrics to proto format",
		zap.Int("metrics_count", len(pbMetrics)))

	req := &pb.UpdateMetricsBatchRequest{Metrics: pbMetrics}

	md := metadata.New(nil)

	if c.secretKey != "" {
		requestData, _ := json.Marshal(req)
		hashValue := hash.CalculateHash(string(requestData), c.secretKey)
		md.Set("hashsha256", hashValue)
		logger.Log.Debug("Added hash to metadata")
	}

	if c.encryptor != nil {
		md.Set("x-encrypted", "true")
		logger.Log.Debug("Added encryption flag to metadata")
	}

	ctx = metadata.NewOutgoingContext(ctx, md)

	start := time.Now()
	_, err := c.client.UpdateMetricsBatch(ctx, req)
	duration := time.Since(start)

	if err != nil {
		logger.Log.Error("Failed to send metrics batch via gRPC",
			zap.Error(err),
			zap.Duration("duration", duration))
	} else {
		logger.Log.Info("Metrics batch sent successfully via gRPC",
			zap.Duration("duration", duration),
			zap.Int("metrics_count", len(pbMetrics)))
	}

	return err
}

// convertToProtoMetrics converts internal metrics to proto metrics
func convertToProtoMetrics(metrics interface{}) []*pb.Metric {
	var pbMetrics []*pb.Metric

	if m, ok := metrics.(*agentmetrics.Metrics); ok {
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "Alloc",
			Type:  "gauge",
			Value: m.Alloc,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "BuckHashSys",
			Type:  "gauge",
			Value: m.BuckHashSys,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "Frees",
			Type:  "gauge",
			Value: m.Frees,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "GCCPUFraction",
			Type:  "gauge",
			Value: m.GCCPUFraction,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "GCSys",
			Type:  "gauge",
			Value: m.GCSys,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "HeapAlloc",
			Type:  "gauge",
			Value: m.HeapAlloc,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "HeapIdle",
			Type:  "gauge",
			Value: m.HeapIdle,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "HeapInuse",
			Type:  "gauge",
			Value: m.HeapInuse,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "HeapObjects",
			Type:  "gauge",
			Value: m.HeapObjects,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "HeapReleased",
			Type:  "gauge",
			Value: m.HeapReleased,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "HeapSys",
			Type:  "gauge",
			Value: m.HeapSys,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "LastGC",
			Type:  "gauge",
			Value: m.LastGC,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "Lookups",
			Type:  "gauge",
			Value: m.Lookups,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "MCacheInuse",
			Type:  "gauge",
			Value: m.MCacheInuse,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "MCacheSys",
			Type:  "gauge",
			Value: m.MCacheSys,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "MSpanInuse",
			Type:  "gauge",
			Value: m.MSpanInuse,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "MSpanSys",
			Type:  "gauge",
			Value: m.MSpanSys,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "Mallocs",
			Type:  "gauge",
			Value: m.Mallocs,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "NextGC",
			Type:  "gauge",
			Value: m.NextGC,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "NumForcedGC",
			Type:  "gauge",
			Value: m.NumForcedGC,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "NumGC",
			Type:  "gauge",
			Value: m.NumGC,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "OtherSys",
			Type:  "gauge",
			Value: m.OtherSys,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "PauseTotalNs",
			Type:  "gauge",
			Value: m.PauseTotalNs,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "StackInuse",
			Type:  "gauge",
			Value: m.StackInuse,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "StackSys",
			Type:  "gauge",
			Value: m.StackSys,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "Sys",
			Type:  "gauge",
			Value: m.Sys,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "TotalAlloc",
			Type:  "gauge",
			Value: m.TotalAlloc,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "RandomValue",
			Type:  "gauge",
			Value: m.RandomValue,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "TotalMemory",
			Type:  "gauge",
			Value: m.TotalMemory,
		})
		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "FreeMemory",
			Type:  "gauge",
			Value: m.FreeMemory,
		})

		for i, cpuPercent := range m.CPUUtilization {
			pbMetrics = append(pbMetrics, &pb.Metric{
				Id:    fmt.Sprintf("CPUUtilization_%d", i),
				Type:  "gauge",
				Value: cpuPercent,
			})
		}

		pbMetrics = append(pbMetrics, &pb.Metric{
			Id:    "PollCount",
			Type:  "counter",
			Delta: m.PollCount,
		})
	}

	return pbMetrics
}

// SendMetric sends a single metric
func (c *Client) SendMetric(ctx context.Context, metric interface{}) error {
	logger.Log.Info("Sending single metric via gRPC")

	pbMetric := convertToProtoMetric(metric)

	req := &pb.UpdateMetricRequest{Metric: pbMetric}

	md := metadata.New(nil)
	if c.secretKey != "" {
		requestData, _ := json.Marshal(req)
		hashValue := hash.CalculateHash(string(requestData), c.secretKey)
		md.Set("hashsha256", hashValue)
	}

	if c.encryptor != nil {
		md.Set("x-encrypted", "true")
	}

	ctx = metadata.NewOutgoingContext(ctx, md)

	start := time.Now()
	_, err := c.client.UpdateMetric(ctx, req)
	duration := time.Since(start)

	if err != nil {
		logger.Log.Error("Failed to send single metric via gRPC",
			zap.Error(err),
			zap.Duration("duration", duration))
	} else {
		logger.Log.Info("Single metric sent successfully via gRPC",
			zap.Duration("duration", duration))
	}

	return err
}

// convertToProtoMetric converts a single internal metric to proto metric
func convertToProtoMetric(metric interface{}) *pb.Metric {
	if m, ok := metric.(*agentmetrics.Metrics); ok {
		return &pb.Metric{
			Id:    "Alloc",
			Type:  "gauge",
			Value: m.Alloc,
		}
	}
	return &pb.Metric{}
}
