package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
	"github.com/lks-go/yandex-praktikum-diploma/internal/service/auth"
)

type TokenParser interface {
	ParseToken(token string) (*auth.Claims, error)
}

func New(tp TokenParser) *Middleware {
	return &Middleware{
		tokenParser: tp,
	}
}

type Middleware struct {
	tokenParser TokenParser
}

func (mw *Middleware) CheckAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		if mustSkip(r) {
			next.ServeHTTP(w, r)
			return
		}

		var login string
		var claims *auth.Claims

		cookie, err := r.Cookie(auth.CookieName)
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				w.WriteHeader(http.StatusUnauthorized)
				return
			default:
				log.Println("cookie error:", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
				return
			}
		}

		claims, err = mw.tokenParser.ParseToken(cookie.Value)
		if err != nil && !errors.Is(err, service.ErrInvalidToken) && !errors.Is(err, service.ErrTokenExpired) {
			log.Println("failed to parse jwt:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}

		if errors.Is(err, service.ErrTokenExpired) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if claims != nil && claims.Login == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if claims != nil {
			login = claims.Login
		}

		r.Header.Set(auth.LoginHeaderName, login)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

var skipURls = []struct {
	method string
	path   string
}{
	{method: http.MethodPost, path: "/api/user/register"},
	{method: http.MethodPost, path: "/api/user/login"},
}

func mustSkip(r *http.Request) bool {

	for _, skip := range skipURls {
		if skip.method == r.Method && strings.Contains(r.URL.Path, skip.path) {
			return true
		}
	}

	return false
}
