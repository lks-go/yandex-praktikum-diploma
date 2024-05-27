package middleware

import (
	"errors"
	"log"
	"net/http"

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
