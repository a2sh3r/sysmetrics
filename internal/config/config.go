package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	PollInterval   float64 `env:"POLL_INTERVAL" envDefault:"2"`
	ReportInterval float64 `env:"REPORT_INTERVAL" envDefault:"10"`
	Address        string  `env:"ADDRESS" envDefault:"localhost:8080"`
}

type ServerConfig struct {
	Address         string `env:"ADDRESS" envDefault:"localhost:8080"`
	LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`
	StoreInterval   int    `env:"STORE_INTERVAL" envDefault:"300"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"/tmp/metrics-db.json"`
	Restore         bool   `env:"RESTORE" envDefault:"true"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:"postgres://postgres:postgres@localhost:5432/postgres"`
}

func NewAgentConfig() (*AgentConfig, error) {
	cfg := &AgentConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	cfg.Address = "http://" + cfg.Address

	return cfg, nil
}

func NewServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return cfg, nil
}
