package service

import (
	"errors"
)

var (
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrInvalidToken          = errors.New("invalid token")
	ErrTokenExpired          = errors.New("token expired")
	ErrUsersPasswordNotMatch = errors.New("users password not match")
	ErrUserNotFound          = errors.New("user not found")
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderAlreadyExists    = errors.New("order already exists")
	ErrOrderConflict         = errors.New("order already exists")

	ErrThirdPartyOrderNotRegistered = errors.New("third party order not registered")
	ErrThirdPartyToManyRequests     = errors.New("third party to many requests")
	ErrThirdPartyInternal           = errors.New("third party internal error")
)

type ErrAuth struct {
	Err  error
	Desc string
}

func (e ErrAuth) Error() string {
	return e.Err.Error()
}

type User struct {
	ID           string
	Login        string
	PasswordHash string
}

type OrderStatus string

const (
	Registered OrderStatus = "REGISTERED"
	Invalid    OrderStatus = "INVALID"
	Processing OrderStatus = "PROCESSING"
	Processed  OrderStatus = "PROCESSED"
)

type Order struct {
	ID         string
	UserID     string
	Number     string
	Status     OrderStatus
	Accrual    int
	UploadedAt string
}

type OrderEvent struct {
	OrderNumber string
	Attempt     int32
}
