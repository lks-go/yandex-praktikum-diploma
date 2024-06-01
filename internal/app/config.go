package app

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DefaultServerAddress = ":8000"

type Config struct {
	NetAddress                NetAddress
	DatabaseDSN               string
	HttpServerShutdownTimeout time.Duration
	AccrualSystemAddress      string
}

func (a *app) BuildConfig() (Config, error) {
	cfg := Config{
		HttpServerShutdownTimeout: time.Second * 2,
	}
	flag.Var(&cfg.NetAddress, "a", "Net address host:port")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database connection string")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system address")
	flag.Parse()

	return cfg, nil
}

type NetAddress struct {
	Host string
	Port int
}

func (a *NetAddress) String() string {
	if a.Port == 0 {
		return DefaultServerAddress
	}

	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *NetAddress) Set(s string) error {
	addr := strings.Split(s, ":")

	if len(addr) < 2 {
		return errors.New("invalid address value")
	}

	p, err := strconv.Atoi(addr[1])
	if err != nil {
		return fmt.Errorf("failed to parse port: %w", err)
	}

	a.Host = addr[0]
	a.Port = p

	return nil
}
