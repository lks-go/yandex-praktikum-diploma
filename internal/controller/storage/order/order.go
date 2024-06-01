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
	q := `SELECT id, user_id, order_number, status, accrual  FROM orders WHERE order_number = $1;`

	o := order{}
	err := s.db.QueryRowContext(ctx, q, orderNumber).Scan(&o.ID, &o.UserID, &o.Number, &o.Status, &o.Accrual)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, service.ErrNotFound
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

func (s *Storage) AddOrder(ctx context.Context, o *service.Order) (string, error) {
	q := `INSERT INTO orders (user_id, order_number, status, accrual) VALUES ($1, $2, $3, $4) RETURNING id`

	id := ""
	err := s.db.QueryRowContext(ctx, q, o.UserID, o.Number, o.Status, o.Accrual).Scan(&id)
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok {
			if err.Code == pgerrcode.UniqueViolation {
				return "", service.ErrAlreadyExists
			}
		}

		return "", fmt.Errorf("failed to exec query: %w", err)
	}

	return id, nil
}

func (s *Storage) UpdateOrder(ctx context.Context, o *service.Order) error {
	q := `UPDATE orders SET status = $2,  accrual = $3, updated_at = now() WHERE id = $1`

	res, err := s.db.ExecContext(ctx, q, o.ID, o.Status, o.Accrual)
	if err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if cnt == 0 {
		return fmt.Errorf("something went wrong, %d affected row", cnt)
	}

	return nil
}

func (s *Storage) UsersOrders(ctx context.Context, userId string) ([]service.Order, error) {
	q := `SELECT id, user_id, order_number, status, accrual  FROM orders WHERE user_id = $1;`

	row, err := s.db.QueryContext(ctx, q, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, service.ErrNotFound
		}
		return nil, fmt.Errorf("failed to make query: %w", err)
	}
	defer row.Close()

	type ord struct {
		ID      string
		UserID  string
		Number  string
		Status  string
		Accrual int
	}

	orders := make([]service.Order, 0)
	for row.Next() {
		o := ord{}
		if err := row.Scan(&o.ID, &o.UserID, &o.Number, &o.Status, &o.Accrual); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		orders = append(orders, service.Order{
			ID:      o.ID,
			UserID:  o.UserID,
			Number:  o.Number,
			Status:  service.OrderStatus(o.Status),
			Accrual: o.Accrual,
		})
	}

	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("rows fail: %w", err)
	}

	return orders, nil
}
