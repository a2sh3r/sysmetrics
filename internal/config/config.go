// Package config provides configuration structures and functions for the agent and server.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

// Duration is a custom type for parsing time duration from strings
type Duration struct {
	time.Duration
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Duration) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value) * time.Second
	case string:
		duration, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration format: %s", value)
		}
		d.Duration = duration
	default:
		return fmt.Errorf("invalid duration type: %T", value)
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (d *Duration) UnmarshalText(text []byte) error {
	duration, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("invalid duration format: %s", string(text))
	}
	d.Duration = duration
	return nil
}

// MarshalJSON implements json.Marshaler
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// AgentConfig holds configuration for the agent.
type AgentConfig struct {
	RateLimit      int64    `env:"RATE_LIMIT" envDefault:"1" json:"rate_limit,omitempty"`
	PollInterval   Duration `env:"POLL_INTERVAL" envDefault:"2s" json:"poll_interval,omitempty"`
	ReportInterval Duration `env:"REPORT_INTERVAL" envDefault:"10s" json:"report_interval,omitempty"`
	Address        string   `env:"ADDRESS" envDefault:"localhost:8080" json:"address,omitempty"`
	SecretKey      string   `env:"KEY" envDefault:"" json:"key,omitempty"`
	CryptoKey      string   `env:"CRYPTO_KEY" envDefault:"" json:"crypto_key,omitempty"`
}

// ServerConfig holds configuration for the server.
type ServerConfig struct {
	StoreInterval   Duration `env:"STORE_INTERVAL" envDefault:"300s" json:"store_interval,omitempty"`
	Address         string   `env:"ADDRESS" envDefault:"localhost:8080" json:"address,omitempty"`
	LogLevel        string   `env:"LOG_LEVEL" envDefault:"info" json:"log_level,omitempty"`
	FileStoragePath string   `env:"FILE_STORAGE_PATH" envDefault:"/tmp/metrics-db.json" json:"store_file,omitempty"`
	DatabaseDSN     string   `env:"DATABASE_DSN" envDefault:"" json:"database_dsn,omitempty"`
	SecretKey       string   `env:"KEY" envDefault:"" json:"key,omitempty"`
	Restore         bool     `env:"RESTORE" envDefault:"true" json:"restore,omitempty"`
	CryptoKey       string   `env:"CRYPTO_KEY" envDefault:"" json:"crypto_key,omitempty"`
	TrustedSubnet   string   `env:"TRUSTED_SUBNET" envDefault:"" json:"trusted_subnet,omitempty"`
}

func loadConfigFile(configPath string, config interface{}) error {
	if configPath == "" {
		return nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("warning: failed to close config file: %v\n", closeErr)
		}
	}()

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

	if err := fs.Parse(os.Args[1:]); err != nil {
		return ""
	}

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
