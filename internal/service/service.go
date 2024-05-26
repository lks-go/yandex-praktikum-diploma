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

type TokenBuilder interface {
	BuildNewToken(login string) (string, error)
}

type Config struct {
	PassHashSalt string
}

type Deps struct {
	UserStorage  UserStorage
	TokenBuilder TokenBuilder
}

func New(cfg *Config, d *Deps) *Service {
	return &Service{
		cfg:          cfg,
		userStorage:  d.UserStorage,
		tokenBuilder: d.TokenBuilder,
	}
}

type Service struct {
	cfg          *Config
	userStorage  UserStorage
	tokenBuilder TokenBuilder
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

func (s *Service) hashPassword(pass string) string {
	h := sha256.New()
	h.Write([]byte(pass + s.cfg.PassHashSalt))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}
