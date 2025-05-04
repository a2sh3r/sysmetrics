package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port int
}

func (n *NetAddress) String() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

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

func (cfg *AgentConfig) ParseFlags() {
	addr := new(NetAddress)

	var (
		pollInterval   float64
		reportInterval float64
		logLevel       string
	)

	flag.Var(addr, "a", "Net address host:port")
	flag.Float64Var(&pollInterval, "p", 2, "poll interval to collect metrics")
	flag.Float64Var(&reportInterval, "r", 10, "report interval to report metrics to server")
	flag.StringVar(&logLevel, "l", "info", "log level")

	flag.Parse()

	if addr.Port != 0 {
		cfg.Address = "http://" + addr.String()
	}

	if pollInterval > 0 {
		cfg.PollInterval = pollInterval
	}

	if reportInterval > 0 {
		cfg.ReportInterval = reportInterval
	}
}

func (cfg *ServerConfig) ParseFlags() {
	addr := new(NetAddress)

	var (
		storeInterval   int
		fileStoragePath string
		restore         bool
		databaseDSN     string
	)

	flag.Var(addr, "a", "Net address host:port")
	flag.IntVar(&storeInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&fileStoragePath, "f", "/tmp/metrics-db.json", "file path to store metrics")
	flag.BoolVar(&restore, "r", true, "restore metrics on start")
	flag.StringVar(&databaseDSN, "d", "", "Database DSN")

	flag.Parse()

	if addr.Port != 0 {
		cfg.Address = addr.String()
	}

	if envValue, exists := os.LookupEnv("STORE_INTERVAL"); exists {
		val, err := strconv.Atoi(envValue)
		if err != nil {
			return
		}
		cfg.StoreInterval = val
	} else {
		cfg.StoreInterval = storeInterval
	}

	if envValue, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists {
		cfg.FileStoragePath = envValue
	} else {
		cfg.FileStoragePath = fileStoragePath
	}

	if envValue, exists := os.LookupEnv("RESTORE"); exists {
		if value, err := strconv.ParseBool(envValue); err == nil {
			cfg.Restore = value
		}
	} else {
		cfg.Restore = restore
	}

	if databaseDSN != "" {
		cfg.DatabaseDSN = databaseDSN
	} else if envValue, exists := os.LookupEnv("DATABASE_DSN"); exists {
		cfg.DatabaseDSN = envValue
	}
}
