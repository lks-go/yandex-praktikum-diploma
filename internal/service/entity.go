package service

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
)

type ErrAuth struct {
	Err error
}

func (e *ErrAuth) Error() string {
	return e.Err.Error()
}
