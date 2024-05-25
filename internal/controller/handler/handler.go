package handler

import "net/http"

func New() *Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {

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
