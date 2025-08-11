package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
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
		pollInterval   float64
		reportInterval float64
		secretKey      string
		rateLimit      int64
		cryptoKey      string
	)

	fs.Var(addr, "a", "Net address host:port")
	fs.Float64Var(&pollInterval, "p", 2, "poll interval to collect metrics")
	fs.Float64Var(&reportInterval, "r", 10, "report interval to report metrics to server")
	fs.StringVar(&secretKey, "k", "", "secret key to calculate hash")
	fs.Int64Var(&rateLimit, "l", 1, "number of parallel workers")
	fs.StringVar(&cryptoKey, "crypto-key", "", "path to public key file for encryption")

	fs.Parse(os.Args[1:])

	if addr.Port != 0 {
		cfg.Address = "http://" + addr.String()
	}

	if pollInterval > 0 {
		cfg.PollInterval = pollInterval
	}

	if reportInterval > 0 {
		cfg.ReportInterval = reportInterval
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
		storeInterval   int
		fileStoragePath string
		restore         bool
		logLevel        string
		databaseDSN     string
		secretKey       string
		cryptoKey       string
	)

	fs.Var(addr, "a", "Net address host:port")
	fs.IntVar(&storeInterval, "i", 300, "store interval in seconds")
	fs.StringVar(&fileStoragePath, "f", "/tmp/metrics-db.json", "file path to store metrics")
	fs.StringVar(&logLevel, "l", "info", "log level")
	fs.BoolVar(&restore, "r", true, "restore metrics on start")
	fs.StringVar(&databaseDSN, "d", "", "Database DSN")
	fs.StringVar(&secretKey, "k", "", "secret key to calculate hash")
	fs.StringVar(&cryptoKey, "crypto-key", "", "path to private key file for decryption")

	fs.Parse(os.Args[1:])

	if addr.Port != 0 {
		cfg.Address = addr.String()
	}

	if storeInterval > 0 {
		cfg.StoreInterval = storeInterval
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
