package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
)

func New(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

type Storage struct {
	db *sql.DB
}

func (s *Storage) AddUser(ctx context.Context, login string, password string) (string, error) {
	q := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`

	id := ""
	err := s.db.QueryRowContext(ctx, q, login, password).Scan(&id)
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok {
			if err.Code == pgerrcode.UniqueViolation {
				return "", service.ErrUserAlreadyExists
			}
		}

		return "", fmt.Errorf("failed to exec query: %w", err)
	}

	return id, nil
}
