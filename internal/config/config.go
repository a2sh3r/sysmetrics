package config

import (
	"time"
)

type AgentConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	Address        string
}

type ServerConfig struct {
	Address string
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
		Address: ":8080",
	}
}
