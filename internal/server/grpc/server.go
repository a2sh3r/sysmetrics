// Package grpc provides gRPC server functionality for handling metrics requests.
package grpc

import (
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/a2sh3r/sysmetrics/internal/config"
	"github.com/a2sh3r/sysmetrics/internal/crypto"
	grpcinterceptors "github.com/a2sh3r/sysmetrics/internal/server/grpc/interceptors"
	pb "github.com/a2sh3r/sysmetrics/proto/metrics"
)

// Server represents the gRPC server
type Server struct {
	server *grpc.Server
	addr   string
}

// NewServer creates a new gRPC server with interceptors
func NewServer(cfg *config.ServerConfig, handler pb.MetricsServiceServer) *Server {
	var interceptors []grpc.UnaryServerInterceptor

	interceptors = append(interceptors, grpcinterceptors.LoggingInterceptor())

	if cfg.SecretKey != "" {
		interceptors = append(interceptors, grpcinterceptors.AuthInterceptor(cfg.SecretKey))
	}

	if cfg.TrustedSubnet != "" {
		interceptors = append(interceptors, grpcinterceptors.IPCheckInterceptor(cfg.TrustedSubnet))
	}

	if cfg.CryptoKey != "" {
		decryptor, err := crypto.NewDecryptor(cfg.CryptoKey)
		if err == nil {
			interceptors = append(interceptors, grpcinterceptors.DecryptInterceptor(decryptor))
		}
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptors...),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			MaxConnectionAge:  10 * time.Minute,
			Time:              30 * time.Second,
			Timeout:           3 * time.Second,
		}),
	}

	srv := grpc.NewServer(opts...)
	pb.RegisterMetricsServiceServer(srv, handler)
	reflection.Register(srv)

	return &Server{
		server: srv,
		addr:   cfg.GRPCAddress,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	return s.server.Serve(lis)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	s.server.GracefulStop()
}
