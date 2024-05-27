package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

type UserStorage interface {
	UserByLogin(ctx context.Context, login string) (*User, error)
	AddUser(ctx context.Context, login string, passwordHash string) (string, error)
}

type OrderStorage interface {
	OrderByNumber(ctx context.Context, orderNumber string) (order *Order, err error)
	AddOrder(ctx context.Context, o Order) (oderID string, err error)
}

type OrderProcessPublisher interface {
	Publish(ctx context.Context, msg OrderEvent)
}

type TokenBuilder interface {
	BuildNewToken(login string) (string, error)
}

type Config struct {
	PassHashSalt string
}

type Deps struct {
	UserStorage  UserStorage
	OrderStorage OrderStorage
	TokenBuilder TokenBuilder
}

func New(cfg *Config, d *Deps) *Service {
	return &Service{
		cfg:          cfg,
		userStorage:  d.UserStorage,
		orderStorage: d.OrderStorage,
		tokenBuilder: d.TokenBuilder,
	}
}

type Service struct {
	cfg                   *Config
	userStorage           UserStorage
	orderStorage          OrderStorage
	tokenBuilder          TokenBuilder
	orderProcessPublisher OrderProcessPublisher
}

func (s *Service) RegisterUser(ctx context.Context, login string, password string) (string, error) {
	_, err := s.userStorage.AddUser(ctx, login, s.hashPassword(password))
	if err != nil {
		return "", fmt.Errorf("failed to add user to storage: %w", err)
	}

	authToken, err := s.tokenBuilder.BuildNewToken(login)
	if err != nil {
		return "", fmt.Errorf("failed to build token: %w", ErrAuth{Err: err})
	}

	return authToken, nil
}

func (s *Service) AuthUser(ctx context.Context, login string, password string) (string, error) {
	u, err := s.userStorage.UserByLogin(ctx, login)
	if err != nil {
		return "", fmt.Errorf("failed to get user by login: %w", err)
	}

	if u == nil {
		return "", errors.New("something went wrong: variable 'u' must not me nil")
	}

	if u.PasswordHash != s.hashPassword(password) {
		return "", ErrUsersPasswordNotMatch
	}

	authToken, err := s.tokenBuilder.BuildNewToken(login)
	if err != nil {
		return "", fmt.Errorf("failed to build token: %w", ErrAuth{Err: err})
	}

	return authToken, nil
}

func (s *Service) SaveOrder(ctx context.Context, login string, orderNumber string) error {
	user, err := s.userStorage.UserByLogin(ctx, login)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return ErrUserNotFound
		default:
			return fmt.Errorf("failed to get user by login: %w", err)
		}
	}

	if user == nil {
		return fmt.Errorf("something went wrong: user is empty")
	}

	order, err := s.orderStorage.OrderByNumber(ctx, orderNumber)
	if err != nil && !errors.Is(err, ErrOrderNotFound) {
		return fmt.Errorf("failed to get order by number: %w", err)
	}

	if order != nil {
		switch {
		case order.UserID == user.ID:
			return ErrOrderAlreadyExists
		case order.UserID != user.ID:
			return ErrOrderConflict
		}
	}

	newOrder := Order{
		UserID: user.ID,
		Number: orderNumber,
		Status: Registered,
	}

	orderID, err := s.orderStorage.AddOrder(ctx, newOrder)
	if err != nil {
		return fmt.Errorf("failed to add order: %w", err)
	}

	go s.orderProcessPublisher.Publish(ctx, OrderEvent{
		Login:   login,
		OrderID: orderID,
	})

	return nil
}

func (s *Service) hashPassword(pass string) string {
	h := sha256.New()
	h.Write([]byte(pass + s.cfg.PassHashSalt))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}
