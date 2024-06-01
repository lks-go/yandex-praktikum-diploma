package operations

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

func (s *Storage) Current(ctx context.Context, userID string) (float64, error) {
	q := `SELECT COALESCE(sum(amount), 0) FROM operations WHERE user_id = $1`

	var amount float64
	if err := s.db.QueryRowContext(ctx, q, userID).Scan(&amount); err != nil {
		return 0, err
	}

	return amount, nil
}

func (s *Storage) Withdrawn(ctx context.Context, userID string) (float64, error) {
	q := `SELECT COALESCE(sum(amount), 0) FROM operations WHERE user_id = $1 AND amount < 0;`

	var amount float64
	if err := s.db.QueryRowContext(ctx, q, userID).Scan(&amount); err != nil {
		return 0, err
	}

	return amount, nil
}

func (s *Storage) Add(ctx context.Context, o *service.Operation) error {

	q := `INSERT INTO operations (user_id, order_number, amount) VALUES ($1, $2, $3)`

	_, err := s.db.ExecContext(ctx, q, o.UserID, o.OrderNumber, o.Amount)
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok {
			if err.Code == pgerrcode.UniqueViolation {
				return service.ErrAlreadyExists
			}

			return fmt.Errorf("failed to exec query: %w", err)
		}
	}

	return nil
}
