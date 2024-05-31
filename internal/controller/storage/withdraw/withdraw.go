package withdraw

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

func (s *Storage) Withdraw(ctx context.Context, userID string, orderNumber string, amount float64) error {
	q := `INSERT INTO withdraws (user_id, order_number, amount) VALUES ($1, $2, $3);`

	_, err := s.db.ExecContext(ctx, q, userID, orderNumber, amount)
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok {
			if err.Code == pgerrcode.UniqueViolation {
				return service.ErrAlreadyExists
			}
		}

		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}
