package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
)

const (
	CookieName      = "auth_token"
	LoginHeaderName = "Login"
)

type Config struct {
	TokenSecretKey      string
	TokenExpirationTime time.Duration
}

func New(cfg *Config) *Auth {
	return &Auth{
		tokenSecretKey:      cfg.TokenSecretKey,
		tokenExpirationTIme: cfg.TokenExpirationTime,
	}
}

type Auth struct {
	tokenSecretKey      string
	tokenExpirationTIme time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
	Login string `json:"login"`
}

func (t *Auth) BuildNewToken(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.tokenExpirationTIme)),
		},
		Login: login,
	})

	tokenString, err := token.SignedString([]byte(t.tokenSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to get signed string: %w", err)
	}

	return tokenString, nil
}

func (t *Auth) ParseToken(token string) (*Claims, error) {
	claims := Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, &claims, func(jt *jwt.Token) (interface{}, error) {
		if jt.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", jt.Header["alg"])
		}

		return []byte(t.tokenSecretKey), nil
	})
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return nil, fmt.Errorf("failed to parse auth: %w", err)
	}

	if errors.Is(err, jwt.ErrTokenExpired) {
		return nil, service.ErrTokenExpired
	}

	if !parsedToken.Valid {
		return nil, service.ErrInvalidToken
	}

	return &claims, nil
}
