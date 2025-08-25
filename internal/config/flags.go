package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// NetAddress represents a network address with host and port.
type NetAddress struct {
	Host string
	Port int
}

// String returns the string representation of the network address.
func (n *NetAddress) String() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

// Set parses and sets the network address from a string.
func (n *NetAddress) Set(flagValue string) error {
	parts := strings.Split(flagValue, ":")
	if len(parts) != 2 {
		return fmt.Errorf("address must be in format host:port")
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("port must be a number")
	}

	n.Host = parts[0]
	n.Port = port
	return nil
}

// ParseFlags parses command-line flags into the AgentConfig.
// This function should be called after NewAgentConfig() to override values with flags.
func (cfg *AgentConfig) ParseFlags() {
	fs := flag.NewFlagSet("agent", flag.ExitOnError)

	addr := new(NetAddress)

	var (
		configPath     string
		pollInterval   string
		reportInterval string
		secretKey      string
		rateLimit      int64
		cryptoKey      string
	)

	fs.StringVar(&configPath, "c", "", "path to config file")
	fs.StringVar(&configPath, "config", "", "path to config file")
	fs.Var(addr, "a", "Net address host:port")
	fs.StringVar(&pollInterval, "p", "2s", "poll interval (e.g., 2s, 1m, 30s)")
	fs.StringVar(&reportInterval, "r", "10s", "report interval (e.g., 10s, 1m, 30s)")
	fs.StringVar(&secretKey, "k", "", "secret key to calculate hash")
	fs.Int64Var(&rateLimit, "l", 1, "number of parallel workers")
	fs.StringVar(&cryptoKey, "crypto-key", "", "path to public key file for encryption")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return
	}

	if addr.Port != 0 {
		cfg.Address = "http://" + addr.String()
	}

	if pollInterval != "" {
		if duration, err := time.ParseDuration(pollInterval); err == nil {
			cfg.PollInterval = Duration{duration}
		}
	}

	if reportInterval != "" {
		if duration, err := time.ParseDuration(reportInterval); err == nil {
			cfg.ReportInterval = Duration{duration}
		}
	}

	if secretKey != "" {
		cfg.SecretKey = secretKey
	}

	if rateLimit > 0 {
		cfg.RateLimit = rateLimit
	}

	if cryptoKey != "" {
		cfg.CryptoKey = cryptoKey
	}
}

// ParseFlags parses command-line flags into the ServerConfig.
// This function should be called after NewServerConfig() to override values with flags.
func (cfg *ServerConfig) ParseFlags() {
	fs := flag.NewFlagSet("server", flag.ExitOnError)

	addr := new(NetAddress)

	var (
		configPath      string
		storeInterval   string
		fileStoragePath string
		restore         bool
		logLevel        string
		databaseDSN     string
		secretKey       string
		cryptoKey       string
		trustedSubnet   string
	)

	fs.StringVar(&configPath, "c", "", "path to config file")
	fs.StringVar(&configPath, "config", "", "path to config file")
	fs.Var(addr, "a", "Net address host:port")
	fs.StringVar(&storeInterval, "i", "", "store interval (e.g., 300s, 5m, 1h)")
	fs.StringVar(&fileStoragePath, "f", "", "file path to store metrics")
	fs.StringVar(&logLevel, "l", "info", "log level")
	fs.BoolVar(&restore, "r", true, "restore metrics on start")
	fs.StringVar(&databaseDSN, "d", "", "Database DSN")
	fs.StringVar(&secretKey, "k", "", "secret key to calculate hash")
	fs.StringVar(&cryptoKey, "crypto-key", "", "path to private key file for decryption")
	fs.StringVar(&trustedSubnet, "t", "", "trusted subnet CIDR")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return
	}

	if addr.Port != 0 {
		cfg.Address = addr.String()
	}

	if storeInterval != "" {
		if duration, err := time.ParseDuration(storeInterval); err == nil {
			cfg.StoreInterval = Duration{duration}
		}
	}

	if fileStoragePath != "" {
		cfg.FileStoragePath = fileStoragePath
	}

	if logLevel != "" {
		cfg.LogLevel = logLevel
	}

	if fs.Lookup("r") != nil && fs.Lookup("r").Value.String() != "" {
		cfg.Restore = restore
	}

	if databaseDSN != "" {
		cfg.DatabaseDSN = databaseDSN
	}

	if secretKey != "" {
		cfg.SecretKey = secretKey
	}

	if cryptoKey != "" {
		cfg.CryptoKey = cryptoKey
	}

	if trustedSubnet != "" {
		cfg.TrustedSubnet = trustedSubnet
	}

	if cfg.DatabaseDSN != "" && !strings.Contains(cfg.DatabaseDSN, "host=") {
		if strings.HasPrefix(cfg.DatabaseDSN, "postgres://") {
			if strings.HasPrefix(cfg.DatabaseDSN, "postgres:///") {
				cfg.DatabaseDSN = strings.Replace(cfg.DatabaseDSN, "postgres:///", "postgres://localhost/", 1)
			}
		} else {
			cfg.DatabaseDSN = "host=localhost " + cfg.DatabaseDSN
		}
	}
}
