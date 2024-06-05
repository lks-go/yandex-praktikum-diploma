package handler_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lks-go/yandex-praktikum-diploma/internal/controller/handler"
	"github.com/lks-go/yandex-praktikum-diploma/internal/controller/handler/mocks"
	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
	"github.com/lks-go/yandex-praktikum-diploma/internal/service/auth"
)

func TestHandler_RegisterUser(t *testing.T) {

	serviceMock := mocks.NewService(t)
	h := handler.New(logrus.New(), serviceMock)

	cases := []struct {
		name               string
		httpRequest        *http.Request
		mock               func()
		expectedStatusCode int
		expectedCookie     string
	}{
		{
			name: "successful registration",
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/register",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"test","password":"qwerty"}`))),
			),
			mock: func() {
				serviceMock.On("RegisterUser", mock.Anything, "test", "qwerty").
					Return("test-token", nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedCookie:     "auth_token=test-token",
		},
		{
			name: "409 conflict",
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/register",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"test","password":"qwerty"}`))),
			),
			mock: func() {
				serviceMock.On("RegisterUser", mock.Anything, "test", "qwerty").
					Return("", service.ErrAlreadyExists).Once()
			},
			expectedStatusCode: http.StatusConflict,
			expectedCookie:     "",
		},
		{
			name: "400 bad request",
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/register",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"","password":"qwerty"}`))),
			),
			mock:               func() {},
			expectedStatusCode: http.StatusBadRequest,
			expectedCookie:     "",
		},
		{
			name: "500 internal error",
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/register",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"test","password":"qwerty"}`))),
			),
			mock: func() {
				serviceMock.On("RegisterUser", mock.Anything, "test", "qwerty").
					Return("", errors.New("any not declared error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedCookie:     "",
		},
		{
			name: "500 internal error(auth)",
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/register",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"test","password":"qwerty"}`))),
			),
			mock: func() {
				serviceMock.On("RegisterUser", mock.Anything, "test", "qwerty").
					Return("", service.ErrAuth{Err: errors.New("auth error")}).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedCookie:     "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()

			w := httptest.NewRecorder()
			h.RegisterUser(w, tc.httpRequest)

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			if tc.expectedCookie != "" {
				assert.Equal(t, tc.expectedCookie, w.Header().Get("Set-Cookie"))
			}
		})
	}

}

func TestHandler_LoginUser(t *testing.T) {
	serviceMock := mocks.NewService(t)
	h := handler.New(logrus.New(), serviceMock)

	cases := []struct {
		name               string
		httpRequest        *http.Request
		mock               func()
		expectedStatusCode int
		expectedCookie     string
	}{
		{
			name: "successful auth",
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/login",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"test","password":"qwerty"}`))),
			),
			mock: func() {
				serviceMock.On("AuthUser", mock.Anything, "test", "qwerty").
					Return("test-token", nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedCookie:     "auth_token=test-token",
		},
		{
			name: "400 bad request",
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/login",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"test","password":""}`))),
			),
			mock:               func() {},
			expectedStatusCode: http.StatusBadRequest,
			expectedCookie:     "",
		},
		{
			name: "401 unauthorized. ErrNotFound",
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/login",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"test","password":"qwerty"}`))),
			),
			mock: func() {
				serviceMock.On("AuthUser", mock.Anything, "test", "qwerty").
					Return("", service.ErrNotFound).Once()
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedCookie:     "",
		},
		{
			name: "401 unauthorized. ErrUsersPasswordNotMatch",
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/login",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"test","password":"qwerty"}`))),
			),
			mock: func() {
				serviceMock.On("AuthUser", mock.Anything, "test", "qwerty").
					Return("", service.ErrUsersPasswordNotMatch).Once()
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedCookie:     "",
		},
		{
			name: "500 internal. ErrAuth",
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/login",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"test","password":"qwerty"}`))),
			),
			mock: func() {
				serviceMock.On("AuthUser", mock.Anything, "test", "qwerty").
					Return("", service.ErrAuth{}).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedCookie:     "",
		},
		{
			name: "500 internal",
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/login",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"test","password":"qwerty"}`))),
			),
			mock: func() {
				serviceMock.On("AuthUser", mock.Anything, "test", "qwerty").
					Return("", errors.New("any not declared error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedCookie:     "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()

			w := httptest.NewRecorder()
			h.LoginUser(w, tc.httpRequest)

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			if tc.expectedCookie != "" {
				assert.Equal(t, tc.expectedCookie, w.Header().Get("Set-Cookie"))
			}
		})
	}
}

func TestHandler_SaveOrder(t *testing.T) {
	serviceMock := mocks.NewService(t)
	h := handler.New(logrus.New(), serviceMock)

	cases := []struct {
		name               string
		httpRequest        func() *http.Request
		mock               func()
		expectedStatusCode int
	}{
		{
			name: "202 accepted",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodPost,
					"https://test.ru/api/user/orders",
					io.NopCloser(bytes.NewReader([]byte(`9981558796712`))),
				)
				req.Header.Set(auth.LoginHeaderName, "test-user")
				return req
			},
			mock: func() {
				serviceMock.On("SaveOrder", mock.Anything, "test-user", "9981558796712").
					Return(nil).Once()
			},
			expectedStatusCode: http.StatusAccepted,
		},
		{
			name: "422 unprocessable entity",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodPost,
					"https://test.ru/api/user/orders",
					io.NopCloser(bytes.NewReader([]byte(`998155879671223`))),
				)
				req.Header.Set(auth.LoginHeaderName, "test-user")
				return req
			},
			mock:               func() {},
			expectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "200 ok",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodPost,
					"https://test.ru/api/user/orders",
					io.NopCloser(bytes.NewReader([]byte(`9981558796712`))),
				)
				req.Header.Set(auth.LoginHeaderName, "test-user")
				return req
			},
			mock: func() {
				serviceMock.On("SaveOrder", mock.Anything, "test-user", "9981558796712").
					Return(service.ErrAlreadyExists).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "409 conflict",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodPost,
					"https://test.ru/api/user/orders",
					io.NopCloser(bytes.NewReader([]byte(`9981558796712`))),
				)
				req.Header.Set(auth.LoginHeaderName, "test-user")
				return req
			},
			mock: func() {
				serviceMock.On("SaveOrder", mock.Anything, "test-user", "9981558796712").
					Return(service.ErrOrderConflict).Once()
			},
			expectedStatusCode: http.StatusConflict,
		},
		{
			name: "500 internal error",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodPost,
					"https://test.ru/api/user/orders",
					io.NopCloser(bytes.NewReader([]byte(`9981558796712`))),
				)
				req.Header.Set(auth.LoginHeaderName, "test-user")
				return req
			},
			mock: func() {
				serviceMock.On("SaveOrder", mock.Anything, "test-user", "9981558796712").
					Return(errors.New("any unexpected error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()

			w := httptest.NewRecorder()
			h.SaveOrder(w, tc.httpRequest())

			assert.Equal(t, tc.expectedStatusCode, w.Code)
		})
	}

}

func TestHandler_Orders(t *testing.T) {
	serviceMock := mocks.NewService(t)
	h := handler.New(logrus.New(), serviceMock)
	uploadedAtFirst := time.Now().Add(-time.Second * 10)
	uploadedAtSecond := time.Now().Add(-time.Second * 20)

	cases := []struct {
		name               string
		httpRequest        func() *http.Request
		mock               func()
		expectedStatusCode int
		expectedBody       func() string
	}{
		{
			name: "204 no content",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/orders",
					nil,
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-2")
				return req
			},
			mock: func() {
				serviceMock.On("OrderList", mock.Anything, "test-user-2").
					Return(nil, service.ErrNotFound).Once()
			},
			expectedStatusCode: http.StatusNoContent,
			expectedBody:       nil,
		},
		{
			name: "204 no content, second case",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/orders",
					nil,
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-2")
				return req
			},
			mock: func() {
				serviceMock.On("OrderList", mock.Anything, "test-user-2").
					Return(nil, nil).Once()
			},
			expectedStatusCode: http.StatusNoContent,
			expectedBody:       nil,
		},
		{
			name: "200 ok",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/orders",
					nil,
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-2")
				return req
			},
			mock: func() {
				orders := []service.Order{
					{
						ID:         "order-id-1",
						UserID:     "user-id-1",
						Number:     "9981558796712",
						Status:     service.OrderStatusProcessing,
						Accrual:    0,
						UploadedAt: uploadedAtFirst,
					},
					{
						ID:         "order-id-2",
						UserID:     "user-id-3",
						Number:     "9981558796713",
						Status:     service.OrderStatusProcessed,
						Accrual:    13,
						UploadedAt: uploadedAtSecond,
					},
				}
				serviceMock.On("OrderList", mock.Anything, "test-user-2").
					Return(orders, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody: func() string {
				return `[
				{
					"number": "9981558796712",
					"status": "` + string(service.OrderStatusProcessing) + `",
					"uploaded_at": "` + uploadedAtFirst.Format(time.RFC3339) + `"
				},
				{
					"number": "9981558796713",
					"status": "` + string(service.OrderStatusProcessed) + `",
					"accrual": 13,
					"uploaded_at": "` + uploadedAtSecond.Format(time.RFC3339) + `"
				}
			]`
			},
		},
		{
			name: "500 internal error",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/orders",
					nil,
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-2")
				return req
			},
			mock: func() {
				serviceMock.On("OrderList", mock.Anything, "test-user-2").
					Return(nil, errors.New("any unexpected error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()

			w := httptest.NewRecorder()
			h.Orders(w, tc.httpRequest())

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			if tc.expectedBody == nil {
				assert.Equal(t, "", w.Body.String())
			} else {
				assert.JSONEq(t, tc.expectedBody(), w.Body.String())
			}
		})
	}
}

func TestHandler_Balance(t *testing.T) {
	serviceMock := mocks.NewService(t)
	h := handler.New(logrus.New(), serviceMock)

	cases := []struct {
		name               string
		httpRequest        func() *http.Request
		mock               func()
		expectedStatusCode int
		expectedBody       func() string
	}{
		{
			name: "200 ок",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/balance",
					nil,
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-3")
				return req
			},
			mock: func() {
				balance := service.UserBalance{
					Current:   548.34,
					Withdrawn: 201.89,
				}

				serviceMock.On("UserBalance", mock.Anything, "test-user-3").
					Return(&balance, nil).Once()
			},
			expectedStatusCode: http.StatusOK,
			expectedBody: func() string {
				return `{"current": 548.34, "withdrawn": 201.89}`
			},
		},
		{
			name: "500 internal error",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/balance",
					nil,
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-3")
				return req
			},
			mock: func() {
				serviceMock.On("UserBalance", mock.Anything, "test-user-3").
					Return(nil, errors.New("any unexpected error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "500 internal error, 2 case",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/balance",
					nil,
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-3")
				return req
			},
			mock: func() {
				serviceMock.On("UserBalance", mock.Anything, "test-user-3").
					Return(nil, nil).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()

			w := httptest.NewRecorder()
			h.Balance(w, tc.httpRequest())

			assert.Equal(t, tc.expectedStatusCode, w.Code)
			if tc.expectedBody == nil {
				assert.Equal(t, "", w.Body.String())
			} else {
				assert.JSONEq(t, tc.expectedBody(), w.Body.String())
			}
		})
	}
}

func TestHandler_Withdraw(t *testing.T) {
	serviceMock := mocks.NewService(t)
	h := handler.New(logrus.New(), serviceMock)

	cases := []struct {
		name               string
		httpRequest        func() *http.Request
		mock               func()
		expectedStatusCode int
	}{
		{
			name: "200 ok",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/withdraw",
					io.NopCloser(bytes.NewReader([]byte(`{"order": "9981558796712", "sum": 13.51}`))),
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-3")
				return req
			},
			mock: func() {
				serviceMock.On("WithdrawBonuses", mock.Anything, "test-user-3", "9981558796712", float32(13.51)).
					Return(nil).Once()
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "422 unprocessable entity",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/withdraw",
					io.NopCloser(bytes.NewReader([]byte(`{"order": "9981558796712"}`))),
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-3")
				return req
			},
			mock:               func() {},
			expectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "422 unprocessable entity, case 2",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/withdraw",
					io.NopCloser(bytes.NewReader([]byte(`{"sum": 13.51}`))),
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-3")
				return req
			},
			mock:               func() {},
			expectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "402 payment required",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/withdraw",
					io.NopCloser(bytes.NewReader([]byte(`{"order": "9981558796712", "sum": 1013.51}`))),
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-3")
				return req
			},
			mock: func() {
				serviceMock.On("WithdrawBonuses", mock.Anything, "test-user-3", "9981558796712", float32(1013.51)).
					Return(service.ErrNotEnoughBonuses).Once()
			},
			expectedStatusCode: http.StatusPaymentRequired,
		},
		{
			name: "500 internal error",
			httpRequest: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"https://test.ru/api/api/user/withdraw",
					io.NopCloser(bytes.NewReader([]byte(`{"order": "9981558796712", "sum": 1013.51}`))),
				)
				req.Header.Set(auth.LoginHeaderName, "test-user-3")
				return req
			},
			mock: func() {
				serviceMock.On("WithdrawBonuses", mock.Anything, "test-user-3", "9981558796712", float32(1013.51)).
					Return(errors.New("any unexpected error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()

			w := httptest.NewRecorder()
			h.Withdraw(w, tc.httpRequest())

			assert.Equal(t, tc.expectedStatusCode, w.Code)
		})
	}

}
