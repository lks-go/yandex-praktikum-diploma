package handler_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
