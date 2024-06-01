package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type UserStorage interface {
	UserByLogin(ctx context.Context, login string) (*User, error)
	AddUser(ctx context.Context, login string, passwordHash string) (string, error)
}

type OrderStorage interface {
	OrderByNumber(ctx context.Context, orderNumber string) (order *Order, err error)
	AddOrder(ctx context.Context, o *Order) (oderID string, err error)
	UpdateOrder(ctx context.Context, o *Order) error
	UsersOrders(ctx context.Context, userID string) ([]Order, error)
}

type OperationsStorage interface {
	Current(ctx context.Context, userID string) (float32, error)
	Withdrawn(ctx context.Context, userID string) (float32, error)
	Add(ctx context.Context, o *Operation) error
	Withdrawals(ctx context.Context, userID string) ([]Withdrawal, error)
}

type OrderProcessPublisher interface {
	Publish(ctx context.Context, msg OrderEvent)
}

type TokenBuilder interface {
	BuildNewToken(login string) (string, error)
}

type Calculator interface {
	Accrual(ctx context.Context, orderNumber string) (*Order, error)
}

type Config struct {
	PassHashSalt      string
	MaxRepublishCount int32
	RepublishWaitTime time.Duration
}

type Deps struct {
	Log                   *logrus.Logger
	UserStorage           UserStorage
	OrderStorage          OrderStorage
	OperationsStorage     OperationsStorage
	TokenBuilder          TokenBuilder
	OrderProcessPublisher OrderProcessPublisher
	Calculator            Calculator
}

func New(cfg *Config, d *Deps) *Service {
	if cfg.MaxRepublishCount <= 0 {
		cfg.MaxRepublishCount = 3
	}

	if cfg.RepublishWaitTime <= 0 {
		cfg.RepublishWaitTime = time.Second * 3
	}

	return &Service{
		cfg:                   cfg,
		log:                   d.Log,
		userStorage:           d.UserStorage,
		orderStorage:          d.OrderStorage,
		operationsStorage:     d.OperationsStorage,
		tokenBuilder:          d.TokenBuilder,
		orderProcessPublisher: d.OrderProcessPublisher,
		calculator:            d.Calculator,
	}
}

type Service struct {
	cfg                   *Config
	log                   *logrus.Logger
	userStorage           UserStorage
	orderStorage          OrderStorage
	operationsStorage     OperationsStorage
	tokenBuilder          TokenBuilder
	orderProcessPublisher OrderProcessPublisher
	calculator            Calculator
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
		return fmt.Errorf("failed to get user by login[%s]: %w", login, err)
	}

	if user == nil {
		return fmt.Errorf("something went wrong: user is empty")
	}

	order, err := s.orderStorage.OrderByNumber(ctx, orderNumber)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("failed to get order by number: %w", err)
	}

	if order != nil {
		switch {
		case order.UserID == user.ID:
			return ErrAlreadyExists
		case order.UserID != user.ID:
			return ErrOrderConflict
		}
	}

	newOrder := Order{
		UserID: user.ID,
		Number: orderNumber,
		Status: Registered,
	}

	orderID, err := s.orderStorage.AddOrder(ctx, &newOrder)
	if err != nil {
		return fmt.Errorf("failed to add order: %w", err)
	}

	go s.orderProcessPublisher.Publish(ctx, OrderEvent{
		UserID:      user.ID,
		OrderID:     orderID,
		OrderNumber: newOrder.Number,
	})

	return nil
}

func (s *Service) OrderAccrual(ctx context.Context, event OrderEvent) error {
	needRepublish := false

	accOrder, err := s.calculator.Accrual(ctx, event.OrderNumber)
	if err != nil {
		switch {
		case errors.Is(err, ErrThirdPartyOrderNotRegistered),
			errors.Is(err, ErrThirdPartyToManyRequests),
			errors.Is(err, ErrThirdPartyInternal):
			needRepublish = true

			s.log.Errorf("failed to accrual order: %s", err)
		default:
			return fmt.Errorf("failed to accrual order: %w", err)
		}
	}

	if accOrder == nil {
		accOrder = &Order{
			ID: event.OrderID,
		}
	} else {
		accOrder.ID = event.OrderID
	}

	if needRepublish {
		switch {
		case event.Attempt >= s.cfg.MaxRepublishCount:
			s.log.Errorf("event republish limit is over")
			accOrder.Status = Invalid
		default:
			go func() {
				s.log.Printf("republishing event with order[%s], attempt %d", event.OrderNumber, event.Attempt)
				time.Sleep(s.cfg.RepublishWaitTime * time.Duration(event.Attempt))
				event.Attempt++
				s.orderProcessPublisher.Publish(ctx, event)
			}()

			return nil
		}
	}

	if err := s.orderStorage.UpdateOrder(ctx, accOrder); err != nil {
		return fmt.Errorf("failed to update order [number %s]: %w", accOrder.Number, err)
	}

	if accOrder.Accrual > 0 {
		o := Operation{
			UserID:      event.UserID,
			OrderNumber: event.OrderNumber,
			Amount:      accOrder.Accrual,
		}
		if err := s.operationsStorage.Add(ctx, &o); err != nil {
			return fmt.Errorf("failed to add operation: %w", err)
		}
	}

	return nil
}

func (s *Service) OrderList(ctx context.Context, login string) ([]Order, error) {
	user, err := s.userStorage.UserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by login [%s]: %w", login, err)
	}

	orders, err := s.orderStorage.UsersOrders(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, nil
}

func (s *Service) UserBalance(ctx context.Context, login string) (*UserBalance, error) {
	user, err := s.userStorage.UserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by login [%s]: %w", login, err)
	}

	current, err := s.operationsStorage.Current(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user's current balance: %w", err)
	}

	withdrawn, err := s.operationsStorage.Withdrawn(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user's current balance: %w", err)
	}

	ub := UserBalance{
		Current:   current,
		Withdrawn: withdrawn,
	}

	return &ub, nil
}

func (s *Service) WithdrawBonuses(ctx context.Context, login string, orderNumber string, amount float32) error {
	user, err := s.userStorage.UserByLogin(ctx, login)
	if err != nil {
		return fmt.Errorf("failed to get user by login [%s]: %w", login, err)
	}

	currentBalance, err := s.operationsStorage.Current(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get user's current balance: %w", err)
	}

	if currentBalance < amount {
		return ErrNotEnoughBonuses
	}

	op := Operation{
		UserID:      user.ID,
		OrderNumber: orderNumber,
		Amount:      -amount,
	}
	if err := s.operationsStorage.Add(ctx, &op); err != nil {
		return fmt.Errorf("failed to withdraw bonuses: %w", err)
	}

	return nil
}

func (s *Service) Withdrawals(ctx context.Context, login string) ([]Withdrawal, error) {
	user, err := s.userStorage.UserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by login [%s]: %w", login, err)
	}

	withdrawals, err := s.operationsStorage.Withdrawals(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	return withdrawals, err
}

func (s *Service) hashPassword(pass string) string {
	h := sha256.New()
	h.Write([]byte(pass + s.cfg.PassHashSalt))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}
