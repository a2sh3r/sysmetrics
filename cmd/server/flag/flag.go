package flag

import (
	"errors"
	"flag"
	"fmt"
	"github.com/a2sh3r/sysmetrics/internal/config"
	"strconv"
	"strings"
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
	} else if len(address) == 1 {
		if na.Port, err = strconv.Atoi(address[0]); err != nil {
			return errors.New("cant parse port")
		}
		na.Host = "localhost"
	}
	return nil
}

func ParseFlags(cfg *config.ServerConfig) {
	addr := new(NetAddress)

	flag.Var(addr, "a", "Net address host:port")
	flag.Parse()

	if addr.Port != 0 {
		cfg.Address = addr.String()
	}
}
