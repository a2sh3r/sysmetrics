// Package config provides configuration structures and functions for the agent and server.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

// AgentConfig holds configuration for the agent.
type AgentConfig struct {
	RateLimit      int64   `env:"RATE_LIMIT" envDefault:"1" json:"rate_limit,omitempty"`
	PollInterval   float64 `env:"POLL_INTERVAL" envDefault:"2" json:"poll_interval,omitempty"`
	ReportInterval float64 `env:"REPORT_INTERVAL" envDefault:"10" json:"report_interval,omitempty"`
	Address        string  `env:"ADDRESS" envDefault:"localhost:8080" json:"address,omitempty"`
	SecretKey      string  `env:"KEY" envDefault:"" json:"key,omitempty"`
	CryptoKey      string  `env:"CRYPTO_KEY" envDefault:"" json:"crypto_key,omitempty"`
}

// ServerConfig holds configuration for the server.
type ServerConfig struct {
	StoreInterval   int    `env:"STORE_INTERVAL" envDefault:"300" json:"store_interval,omitempty"`
	Address         string `env:"ADDRESS" envDefault:"localhost:8080" json:"address,omitempty"`
	LogLevel        string `env:"LOG_LEVEL" envDefault:"info" json:"log_level,omitempty"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"/tmp/metrics-db.json" json:"store_file,omitempty"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:"" json:"database_dsn,omitempty"`
	SecretKey       string `env:"KEY" envDefault:"" json:"key,omitempty"`
	Restore         bool   `env:"RESTORE" envDefault:"true" json:"restore,omitempty"`
	CryptoKey       string `env:"CRYPTO_KEY" envDefault:"" json:"crypto_key,omitempty"`
}

func loadConfigFile(configPath string, config interface{}) error {
	if configPath == "" {
		return nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	return nil
}

func getConfigPath() string {
	if envPath := os.Getenv("CONFIG"); envPath != "" {
		return envPath
	}

	var configPath string
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.StringVar(&configPath, "c", "", "path to config file")
	fs.StringVar(&configPath, "config", "", "path to config file")

	fs.Parse(os.Args[1:])

	return configPath
}

// NewAgentConfig creates a new AgentConfig with priority: flags > env > config file.
func NewAgentConfig() (*AgentConfig, error) {
	configPath := getConfigPath()

	cfg := &AgentConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if err := loadConfigFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	if cfg.Address != "" && !isHTTPAddress(cfg.Address) {
		cfg.Address = "http://" + cfg.Address
	}

	return cfg, nil
}

// NewServerConfig creates a new ServerConfig with priority: flags > env > config file.
func NewServerConfig() (*ServerConfig, error) {
	configPath := getConfigPath()

	cfg := &ServerConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if err := loadConfigFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	return cfg, nil
}

func isHTTPAddress(addr string) bool {
	return len(addr) >= 7 && (addr[:7] == "http://" || addr[:8] == "https://")
}
