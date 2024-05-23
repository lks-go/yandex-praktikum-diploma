package app

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func RunMigrations(dsn string, path string) error {
	m, err := migrate.New("file://"+path, dsn)
	if err != nil {
		return fmt.Errorf("feailed to create new: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to up migrations: %w", err)
	}

	return nil
}
