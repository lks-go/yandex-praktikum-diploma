package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
	"github.com/lks-go/yandex-praktikum-diploma/internal/service/auth"
)

type Service interface {
	RegisterUser(ctx context.Context, login string, password string) (string, error)
}

func New(log *logrus.Logger, s Service) *Handler {
	return &Handler{
		log:     log,
		service: s,
	}
}

type Handler struct {
	log     *logrus.Logger
	service Service
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	l := h.log.WithField("handler", "RegisterUser")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		l.Errorf("failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	requestBody := struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}{}

	if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
		l.Errorf("failed to unmarshal request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	errAuth := service.ErrAuth{}

	authToken, err := h.service.RegisterUser(r.Context(), requestBody.Login, requestBody.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists):
			w.WriteHeader(http.StatusConflict)
			return
		case errors.As(err, &errAuth):
			l.Errorf("failed to login user: %s", errAuth)
			w.WriteHeader(http.StatusInternalServerError)
			return
		default:
			l.Errorf("failed to create new user: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	newCookie := http.Cookie{
		Name:  auth.CookieName,
		Value: authToken,
	}

	http.SetCookie(w, &newCookie)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) SaveOrder(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) Orders(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) Balance(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) Withdrawals(w http.ResponseWriter, r *http.Request) {

}
