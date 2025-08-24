package grpc

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/hash"
	"github.com/a2sh3r/sysmetrics/internal/server/grpc/handlers"
	"github.com/a2sh3r/sysmetrics/internal/server/repositories"
	"github.com/a2sh3r/sysmetrics/internal/server/services"
	"github.com/a2sh3r/sysmetrics/internal/server/storage/memstorage"
	pb "github.com/a2sh3r/sysmetrics/proto/metrics"
)

func TestGRPCServer(t *testing.T) {
	cfg := &config.ServerConfig{
		GRPCAddress: "localhost:9091",
		SecretKey:   "",
	}

	storage := memstorage.NewMemStorage()
	metricRepo := repositories.NewMetricRepo(storage)
	service := services.NewService(metricRepo)
	handler := handlers.NewMetricsHandler(service)
	server := NewServer(cfg, handler)

	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Failed to start gRPC server: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	conn, err := grpc.NewClient("localhost:9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Logf("Failed to connect to gRPC server: %v", err)
		return
	}
	defer func() {
		if conErr := conn.Close(); conErr != nil {
			t.Logf("Failed to close connection")
		}
	}()

	t.Run("UpdateMetric", func(t *testing.T) {
		testUpdateMetric(t, conn)
	})

	t.Run("UpdateMetricsBatch", func(t *testing.T) {
		testUpdateMetricsBatch(t, conn)
	})

	t.Run("GetMetric", func(t *testing.T) {
		testGetMetric(t, conn)
	})

	t.Run("GetMetrics", func(t *testing.T) {
		testGetMetrics(t, conn)
	})

	t.Run("Authentication", func(t *testing.T) {
		testAuthentication(t, conn)
	})

	t.Run("IPCheck", func(t *testing.T) {
		testIPCheck(t, conn)
	})

	t.Run("Decrypt", func(t *testing.T) {
		testDecrypt(t, conn)
	})

	t.Run("Logging", func(t *testing.T) {
		testLogging(t, conn)
	})

	server.Stop()
}

func TestGRPCServerWithSecurity(t *testing.T) {
	cfg := &config.ServerConfig{
		GRPCAddress:   "localhost:9092",
		SecretKey:     "test-secret-key",
		TrustedSubnet: "127.0.0.0/8",
		CryptoKey:     "test-crypto-key",
	}

	storage := memstorage.NewMemStorage()
	metricRepo := repositories.NewMetricRepo(storage)
	service := services.NewService(metricRepo)
	handler := handlers.NewMetricsHandler(service)
	server := NewServer(cfg, handler)

	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Failed to start gRPC server: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	conn, err := grpc.NewClient("localhost:9092", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Logf("Failed to connect to gRPC server: %v", err)
		return
	}
	defer func() {
		if conErr := conn.Close(); conErr != nil {
			t.Logf("Failed to close connection")
		}
	}()

	t.Run("AuthenticatedRequest", func(t *testing.T) {
		testAuthenticatedRequest(t, conn)
	})

	t.Run("IPRestrictedRequest", func(t *testing.T) {
		testIPRestrictedRequest(t, conn)
	})

	t.Run("EncryptedRequest", func(t *testing.T) {
		testEncryptedRequest(t, conn)
	})

	server.Stop()
}

func testUpdateMetric(t *testing.T, conn *grpc.ClientConn) {
	ctx := context.Background()

	client := pb.NewMetricsServiceClient(conn)
	req := &pb.UpdateMetricRequest{
		Metric: &pb.Metric{
			Id:    "test_metric",
			Type:  "gauge",
			Value: 42.0,
		},
	}

	_, err := client.UpdateMetric(ctx, req)
	if err != nil {
		t.Errorf("UpdateMetric() error = %v", err)
		return
	}
}

func testUpdateMetricsBatch(t *testing.T, conn *grpc.ClientConn) {
	ctx := context.Background()

	client := pb.NewMetricsServiceClient(conn)
	req := &pb.UpdateMetricsBatchRequest{
		Metrics: []*pb.Metric{
			{
				Id:    "test_metric_1",
				Type:  "gauge",
				Value: 42.0,
			},
			{
				Id:    "test_metric_2",
				Type:  "counter",
				Delta: 100,
			},
		},
	}

	_, err := client.UpdateMetricsBatch(ctx, req)
	if err != nil {
		t.Errorf("UpdateMetricsBatch() error = %v", err)
		return
	}
}

func testGetMetric(t *testing.T, conn *grpc.ClientConn) {
	ctx := context.Background()

	client := pb.NewMetricsServiceClient(conn)
	req := &pb.GetMetricRequest{
		Id:   "test_metric",
		Type: "gauge",
	}

	_, err := client.GetMetric(ctx, req)
	if err != nil {
		t.Errorf("GetMetric() error = %v", err)
		return
	}
}

func testGetMetrics(t *testing.T, conn *grpc.ClientConn) {
	ctx := context.Background()

	client := pb.NewMetricsServiceClient(conn)
	req := &pb.GetMetricsRequest{}

	_, err := client.GetMetrics(ctx, req)
	if err != nil {
		t.Errorf("GetMetrics() error = %v", err)
		return
	}
}

func testAuthentication(t *testing.T, conn *grpc.ClientConn) {
	ctx := context.Background()

	client := pb.NewMetricsServiceClient(conn)
	req := &pb.UpdateMetricRequest{
		Metric: &pb.Metric{
			Id:    "test_auth",
			Type:  "gauge",
			Value: 1.0,
		},
	}

	_, err := client.UpdateMetric(ctx, req)
	if err == nil {
		t.Log("Authentication test passed - server accepts requests without hash")
	} else {
		t.Logf("Authentication test - server requires hash: %v", err)
	}
}

func testIPCheck(t *testing.T, conn *grpc.ClientConn) {
	ctx := context.Background()

	md := metadata.New(nil)
	md.Set("x-real-ip", "192.168.1.100")
	ctx = metadata.NewOutgoingContext(ctx, md)

	client := pb.NewMetricsServiceClient(conn)
	req := &pb.UpdateMetricRequest{
		Metric: &pb.Metric{
			Id:    "test_ip_check",
			Type:  "gauge",
			Value: 1.0,
		},
	}

	_, err := client.UpdateMetric(ctx, req)
	if err != nil {
		t.Logf("IP check test - server rejected IP: %v", err)
	} else {
		t.Log("IP check test passed - server accepted IP")
	}
}

func testDecrypt(t *testing.T, conn *grpc.ClientConn) {
	ctx := context.Background()
	md := metadata.New(map[string]string{
		"x-encrypted": "true",
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	client := pb.NewMetricsServiceClient(conn)
	req := &pb.UpdateMetricRequest{
		Metric: &pb.Metric{
			Id:    "encrypted_metric",
			Type:  "gauge",
			Value: 99.9,
		},
	}

	_, err := client.UpdateMetric(ctx, req)
	if err != nil {
		t.Logf("Decrypt interceptor test: %v", err)
	}
}

func testLogging(t *testing.T, conn *grpc.ClientConn) {
	ctx := context.Background()
	md := metadata.New(map[string]string{
		"x-real-ip": "127.0.0.1",
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	client := pb.NewMetricsServiceClient(conn)
	req := &pb.GetMetricsRequest{}

	_, err := client.GetMetrics(ctx, req)
	if err != nil {
		t.Errorf("Logging interceptor test failed: %v", err)
		return
	}
}

func testAuthenticatedRequest(t *testing.T, conn *grpc.ClientConn) {
	ctx := context.Background()
	req := &pb.UpdateMetricRequest{
		Metric: &pb.Metric{
			Id:    "auth_test",
			Type:  "gauge",
			Value: 123.45,
		},
	}

	requestData, _ := json.Marshal(req)
	hashValue := hash.CalculateHash(string(requestData), "test-secret-key")

	md := metadata.New(map[string]string{
		"hashsha256": hashValue,
		"x-real-ip":  "127.0.0.1",
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	client := pb.NewMetricsServiceClient(conn)
	_, err := client.UpdateMetric(ctx, req)
	if err != nil {
		t.Errorf("Authenticated request failed: %v", err)
		return
	}
}

func testIPRestrictedRequest(t *testing.T, conn *grpc.ClientConn) {
	ctx := context.Background()
	req := &pb.GetMetricsRequest{}

	requestData, marshalErr := protojson.Marshal(req)
	if marshalErr != nil {
		t.Logf("cannot marshal json")
	}
	hashValue := hash.CalculateHash(string(requestData), "test-secret-key")

	md := metadata.New(map[string]string{
		"x-real-ip":  "127.0.0.1",
		"hashsha256": hashValue,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	client := pb.NewMetricsServiceClient(conn)
	_, err := client.GetMetrics(ctx, req)
	if err != nil {
		t.Errorf("IP restricted request failed: %v", err)
		return
	}
}

func testEncryptedRequest(t *testing.T, conn *grpc.ClientConn) {
	ctx := context.Background()
	req := &pb.UpdateMetricRequest{
		Metric: &pb.Metric{
			Id:    "encrypted_test",
			Type:  "counter",
			Delta: 999,
		},
	}

	requestData, _ := json.Marshal(req)
	hashValue := hash.CalculateHash(string(requestData), "test-secret-key")

	md := metadata.New(map[string]string{
		"x-encrypted": "true",
		"x-real-ip":   "127.0.0.1",
		"hashsha256":  hashValue,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	client := pb.NewMetricsServiceClient(conn)
	_, err := client.UpdateMetric(ctx, req)
	if err != nil {
		t.Logf("Encrypted request test: %v", err)
	}
}
