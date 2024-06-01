package operations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

func (s *Storage) Current(ctx context.Context, userID string) (float32, error) {
	q := `SELECT COALESCE(sum(amount), 0) FROM operations WHERE user_id = $1`

	var amount float32
	if err := s.db.QueryRowContext(ctx, q, userID).Scan(&amount); err != nil {
		return 0, err
	}

	return amount, nil
}

func (s *Storage) Withdrawn(ctx context.Context, userID string) (float32, error) {
	q := `SELECT COALESCE(sum(amount), 0) FROM operations WHERE user_id = $1 AND amount < 0;`

	var amount float32
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

func (s *Storage) Withdrawals(ctx context.Context, userID string) ([]service.Withdrawal, error) {
	q := `SELECT order_number, amount, created_at  FROM operations WHERE user_id = $1 AND amount < 0;`

	row, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, service.ErrNotFound
		}
		return nil, fmt.Errorf("failed to make query: %w", err)
	}
	defer row.Close()

	type withdrawalDTO struct {
		OrderNumber string
		Amount      float64
		CreatedAt   time.Time
	}

	withdrawals := make([]service.Withdrawal, 0)
	for row.Next() {
		dto := withdrawalDTO{}
		if err := row.Scan(&dto.OrderNumber, &dto.Amount, &dto.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}

		withdrawals = append(withdrawals, service.Withdrawal{
			OrderNumber: dto.OrderNumber,
			Amount:      -dto.Amount,
			ProcessedAt: dto.CreatedAt,
		})
	}

	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("rows fail: %w", err)
	}

	return withdrawals, nil
}
