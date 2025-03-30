package config

import (
	"time"
)

type AgentConfig struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	Address        string        `env:"ADDRESS"`
}

type ServerConfig struct {
	Address string `env:"ADDRESS"`
}

func NewAgentConfig() *AgentConfig {
	return &AgentConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		Address:        "http://localhost:8080",
	}
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Address: "localhost:8080",
	}
}
