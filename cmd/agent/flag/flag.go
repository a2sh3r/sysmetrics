package flag

import (
	"errors"
	"flag"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/config"
	"strconv"
	"strings"
	"time"
)

type NetAddress struct {
	Host string
	Port int
}

func (na *NetAddress) String() string {
	return fmt.Sprint(na.Host, ":", strconv.Itoa(na.Port))
}

func (na *NetAddress) Set(flagValue string) error {
	address := strings.Split(flagValue, ":")
	var err error

	if len(address) == 2 {
		na.Host = address[0]
		if na.Port, err = strconv.Atoi(address[1]); err != nil {
			return errors.New("cant parse port")
		}
	} else {
		if na.Port, err = strconv.Atoi(address[0]); err != nil {
			return errors.New("cant parse port")
		}
		na.Host = "localhost"
	}
	return nil
}

func ParseFlags(cfg *config.AgentConfig) {
	addr := new(NetAddress)

	var (
		pollInterval   float64
		reportInterval float64
	)

	flag.Var(addr, "a", "Net address host:port")
	flag.Float64Var(&pollInterval, "p", 2, "poll interval to collect metrics")
	flag.Float64Var(&reportInterval, "r", 10, "report interval to report metrics to server")

	flag.Parse()

	cfg.Address = "http://" + addr.String()
	cfg.PollInterval = time.Duration(pollInterval) * time.Second
	cfg.ReportInterval = time.Duration(reportInterval) * time.Second
}
