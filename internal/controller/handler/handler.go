package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/sirupsen/logrus"

	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
	"github.com/lks-go/yandex-praktikum-diploma/internal/service/auth"
)

type Service interface {
	RegisterUser(ctx context.Context, login string, password string) (token string, err error)
	AuthUser(ctx context.Context, login string, password string) (token string, err error)
	SaveOrder(ctx context.Context, login string, orderNumber string) error
	OrderList(ctx context.Context, login string) ([]service.Order, error)
	UserBalance(ctx context.Context, login string) (*service.UserBalance, error)
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
		Login    *string `json:"login,omitempty"`
		Password *string `json:"password,omitempty"`
	}{}

	if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
		l.Errorf("failed to unmarshal request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	validationErrors := make([]error, 0)

	if err := validateLogin(requestBody.Login); err != nil {
		validationErrors = append(validationErrors, err)
	}

	if err := validatePassword(requestBody.Login); err != nil {
		validationErrors = append(validationErrors, err)
	}

	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			l.Warnf("login not valid: %s", e)
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	errAuth := service.ErrAuth{}

	authToken, err := h.service.RegisterUser(r.Context(), *requestBody.Login, *requestBody.Password)
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
	defer r.Body.Close()
	l := h.log.WithField("handler", "LoginUser")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		l.Errorf("failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	requestBody := struct {
		Login    *string `json:"login,omitempty"`
		Password *string `json:"password,omitempty"`
	}{}

	if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
		l.Errorf("failed to unmarshal request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	validationErrors := make([]error, 0)

	if err := validateLogin(requestBody.Login); err != nil {
		validationErrors = append(validationErrors, err)
	}

	if err := validatePassword(requestBody.Password); err != nil {
		validationErrors = append(validationErrors, err)
	}

	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			l.Warnf("login not valid: %s", e)
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	errAuth := service.ErrAuth{}

	authToken, err := h.service.AuthUser(r.Context(), *requestBody.Login, *requestBody.Password)
	if err != nil {
		switch {
		case errors.As(err, &errAuth):
			l.Errorf("failed to login user: %s", errAuth)
			w.WriteHeader(http.StatusInternalServerError)
			return
		case errors.Is(err, service.ErrUserNotFound):
			l.Warnf("failed to login user: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		case errors.Is(err, service.ErrUsersPasswordNotMatch):
			l.Warn(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			l.Errorf("failed to auth user: %s", err)
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

func (h *Handler) SaveOrder(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	l := h.log.WithField("handler", "SaveOrder")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		l.Errorf("failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	requestBody := string(bodyBytes)
	if err := validateOrderNumber(requestBody); err != nil {
		l.Warnf("invalid order number: %s", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = h.service.SaveOrder(r.Context(), r.Header.Get(auth.LoginHeaderName), requestBody)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderAlreadyExists):
			w.WriteHeader(http.StatusOK)
			return
		case errors.Is(err, service.ErrOrderConflict):
			w.WriteHeader(http.StatusConflict)
			return
		default:
			l.Errorf("failed to save order: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) Orders(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	l := h.log.WithField("handler", "Orders")

	orderList, err := h.service.OrderList(r.Context(), r.Header.Get(auth.LoginHeaderName))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderNotFound):
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			l.Errorf("failed to get order list: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	type orderDTO struct {
		Number     string `json:"number"`
		Status     string `json:"status"`
		Accrual    *int   `json:"accrual,omitempty"`
		UploadedAt string `json:"uploaded_at"`
	}

	orders := make([]orderDTO, 0, len(orderList))

	for _, o := range orderList {
		order := orderDTO{
			Number:     o.Number,
			Status:     string(o.Status),
			UploadedAt: o.UploadedAt,
		}

		if o.Accrual > 0 {
			*order.Accrual = o.Accrual
		}

		orders = append(orders, order)
	}

	body, err := json.Marshal(orders)
	if err != nil {
		l.Errorf("failed to marshal order list: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(body); err != nil {
		l.Errorf("failed to write body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Balance(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	l := h.log.WithField("handler", "Balance")

	userBalance, err := h.service.UserBalance(r.Context(), r.Header.Get(auth.LoginHeaderName))
	if err != nil {
		l.Errorf("failed to get user balance: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if userBalance == nil {
		l.Errorf("something went wrong, userBalance is nil")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type dto struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}

	body, err := json.Marshal(dto(*userBalance))
	if err != nil {
		l.Errorf("failed to marshal order list: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(body); err != nil {
		l.Errorf("failed to write body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) Withdrawals(w http.ResponseWriter, r *http.Request) {

}

func validateLogin(login *string) error {
	if login == nil || *login == "" {
		return fmt.Errorf("login must not be empty")
	}

	return nil
}

func validatePassword(pass *string) error {
	if pass == nil || *pass == "" {
		return fmt.Errorf("password must not be empty")
	}

	return nil
}

func validateOrderNumber(orderNumber string) error {
	err := goluhn.Validate(orderNumber)
	if err != nil {
		return err
	}

	return nil
}
