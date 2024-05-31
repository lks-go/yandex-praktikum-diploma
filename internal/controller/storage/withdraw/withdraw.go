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

func (s *Storage) Withdrawals(ctx context.Context, userID string) ([]service.Withdrawal, error) {
	q := `SELECT order_number, amount, processed_at  FROM orders WHERE user_id = $1;`

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
		ProcessedAt string
	}

	withdrawals := make([]service.Withdrawal, 0)
	for row.Next() {
		dto := withdrawalDTO{}
		if err := row.Scan(&dto.OrderNumber, &dto.Amount, &dto.ProcessedAt); err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}

		withdrawals = append(withdrawals, service.Withdrawal{
			OrderNumber: dto.OrderNumber,
			Amount:      dto.Amount,
			ProcessedAt: dto.ProcessedAt,
		})
	}

	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("rows fail: %w", err)
	}

	return withdrawals, nil
}
