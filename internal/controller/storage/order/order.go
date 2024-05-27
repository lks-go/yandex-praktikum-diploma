package order

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

type order struct {
	ID         string
	UserID     string
	Number     string
	Status     string
	Accrual    int
	UploadedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (s *Storage) OrderByNumber(ctx context.Context, orderNumber string) (*service.Order, error) {
	q := `SELECT id, user_id, order_number, status, accrual  FROM orders WHERE number = $1;`

	o := order{}
	err := s.db.QueryRowContext(ctx, q, orderNumber).Scan(&o.ID, &o.UserID, &o.Number, &o.Status, &o.Accrual)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, service.ErrOrderNotFound
		}

		return nil, fmt.Errorf("query row error: %w", err)
	}

	so := service.Order{
		ID:      o.ID,
		UserID:  o.UserID,
		Number:  o.Number,
		Status:  service.OrderStatus(o.Status),
		Accrual: o.Accrual,
	}

	return &so, nil

}

func (s *Storage) AddOrder(ctx context.Context, o service.Order) (string, error) {
	q := `INSERT INTO orders (user_id, order_number, status, accrual) VALUES ($1, $2) RETURNING id`

	id := ""
	err := s.db.QueryRowContext(ctx, q, o.UserID, o.Number, o.Status, o.Accrual).Scan(&id)
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok {
			if err.Code == pgerrcode.UniqueViolation {
				return "", service.ErrOrderAlreadyExists
			}
		}

		return "", fmt.Errorf("failed to exec query: %w", err)
	}

	return id, nil
}
