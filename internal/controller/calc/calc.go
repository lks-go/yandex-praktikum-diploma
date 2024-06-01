package calc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"gopkg.in/resty.v1"

	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
)

type Config struct {
	HostURL    string
	RetryCount int
}

func NewHTTPClient(cfg *Config) *Client {
	httpClient := &http.Client{}
	client := resty.NewWithClient(httpClient).
		SetHostURL(cfg.HostURL).
		SetRetryCount(cfg.RetryCount)

	return &Client{client}
}

type Client struct {
	*resty.Client
}

func (c *Client) Accrual(ctx context.Context, orderNumber string) (*service.Order, error) {
	res, err := c.R().SetContext(ctx).Get("/api/orders/" + orderNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	switch res.StatusCode() {
	case http.StatusNoContent:
		return nil, service.ErrThirdPartyOrderNotRegistered
	case http.StatusTooManyRequests:
		return nil, service.ErrThirdPartyToManyRequests
	case http.StatusInternalServerError:
		return nil, service.ErrThirdPartyInternal
	}

	type order struct {
		Number  string
		Status  string
		Accrual int
	}

	o := order{}
	if err := json.Unmarshal(res.Body(), &o); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	so := service.Order{
		Number:  o.Number,
		Status:  service.OrderStatus(o.Status),
		Accrual: o.Accrual,
	}

	return &so, nil
}
