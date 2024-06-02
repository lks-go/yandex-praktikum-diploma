package app

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	NetAddress                string
	DatabaseDSN               string
	HTTPServerShutdownTimeout time.Duration
	AccrualSystemAddress      string
}

func (a *app) BuildConfig() (Config, error) {
	cfg := Config{
		HTTPServerShutdownTimeout: time.Second * 2,
	}
	flag.StringVar(&cfg.NetAddress, "a", ":8000", "Net address host:port")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database connection string")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system address")
	flag.Parse()

	if runAddress, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		cfg.NetAddress = runAddress
	}

	if databaseURI, ok := os.LookupEnv("DATABASE_URI"); ok {
		cfg.DatabaseDSN = databaseURI
	}

	if accrualSystemSddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		cfg.AccrualSystemAddress = accrualSystemSddress
	}

	return cfg, nil
}
