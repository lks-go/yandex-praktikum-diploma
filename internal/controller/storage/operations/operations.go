package operations

import (
	"context"
	"database/sql"
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
	q := `SELECT COALESCE(sum(amount), 0) FROM operations WHERE user_id = = $1`

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
