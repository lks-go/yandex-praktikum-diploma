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
)

func TestRegisterUser(t *testing.T) {

	serviceMock := mocks.NewService(t)
	h := handler.New(logrus.New(), serviceMock)

	cases := []struct {
		name               string
		requestBody        string
		httpRequest        *http.Request
		mock               func()
		expectedStatusCode int
		expectedCookie     string
	}{
		{
			name:        "successful registration",
			requestBody: `{"login":"test","password":"qwerty"}`,
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
			name:        "409 conflict",
			requestBody: `{"login":"test","password":"qwerty"}`,
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
			name:        "400 bad request",
			requestBody: `{"login":"","password":"qwerty"}`,
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
			name:        "500 internal error",
			requestBody: `{"login":"test","password":"qwerty"}`,
			httpRequest: httptest.NewRequest(
				http.MethodPost,
				"https://test.ru/api/user/register",
				io.NopCloser(bytes.NewReader([]byte(`{"login":"test","password":"qwerty"}`))),
			),
			mock: func() {
				serviceMock.On("RegisterUser", mock.Anything, "test", "qwerty").
					Return("", errors.New("any not declares error")).Once()
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedCookie:     "",
		},
		{
			name:        "500 internal error(auth)",
			requestBody: `{"login":"test","password":"qwerty"}`,
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
