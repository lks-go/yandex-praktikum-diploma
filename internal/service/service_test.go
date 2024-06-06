package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
	"github.com/lks-go/yandex-praktikum-diploma/internal/service/mocks"
)

type Suit struct {
	suite.Suite

	userStorage           *mocks.UserStorage
	orderStorage          *mocks.OrderStorage
	operationsStorage     *mocks.OperationsStorage
	tokenBuilder          *mocks.TokenBuilder
	orderProcessPublisher *mocks.OrderProcessPublisher
	calculator            *mocks.Calculator

	log           *logrus.Logger
	serviceConfig service.Config
}

func (s *Suit) SetupTest() {
	s.userStorage = mocks.NewUserStorage(s.T())
	s.orderStorage = mocks.NewOrderStorage(s.T())
	s.operationsStorage = mocks.NewOperationsStorage(s.T())
	s.tokenBuilder = mocks.NewTokenBuilder(s.T())
	s.orderProcessPublisher = mocks.NewOrderProcessPublisher(s.T())
	s.calculator = mocks.NewCalculator(s.T())
	s.log = logrus.New()
	s.serviceConfig = service.Config{
		PassHashSalt:      "TEST_SALT",
		MaxRepublishCount: 3,
		RepublishWaitTime: time.Millisecond * 10,
	}
}

func (s *Suit) TestService_RegisterUser_Positive() {
	ctx := context.Background()
	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6InRlc3RfdXNlciIsImlhdCI6MTUxNjIzOTAyMn0.YTH8MZcIu-j5Fw7fr2zi4KB52c1x0P1d2XlUJ7fak1o"

	deps := service.Deps{
		UserStorage:  s.userStorage,
		TokenBuilder: s.tokenBuilder,
	}
	service := service.New(&s.serviceConfig, &deps)

	login := "test_user"
	password := "test_password"

	s.userStorage.On("AddUser", ctx, login, service.HashPassword(password)).
		Return("user-id", nil).Once()

	s.tokenBuilder.On("BuildNewToken", login).Return(testToken, nil).Once()

	authToken, err := service.RegisterUser(ctx, login, password)
	require.NoError(s.T(), err)
	require.Equal(s.T(), testToken, authToken)
}

func (s *Suit) TestService_RegisterUser_NegativeAddUser() {
	ctx := context.Background()

	deps := service.Deps{
		UserStorage: s.userStorage,
	}
	service := service.New(&s.serviceConfig, &deps)

	login := "test_user"
	password := "test_password"

	s.userStorage.On("AddUser", ctx, login, service.HashPassword(password)).
		Return("", errors.New("any unexpected error")).Once()

	authToken, err := service.RegisterUser(ctx, login, password)
	require.Error(s.T(), err)
	require.Equal(s.T(), "", authToken)
}

func (s *Suit) TestService_RegisterUser_NegativeBuildToken() {
	ctx := context.Background()

	deps := service.Deps{
		UserStorage:  s.userStorage,
		TokenBuilder: s.tokenBuilder,
	}
	sv := service.New(&s.serviceConfig, &deps)

	login := "test_user"
	password := "test_password"

	s.userStorage.On("AddUser", ctx, login, sv.HashPassword(password)).
		Return("user-id", nil).Once()

	tokenErr := fmt.Errorf("failed to build token: %w", service.ErrAuth{Err: errors.New("some error")})
	s.tokenBuilder.On("BuildNewToken", login).
		Return("", tokenErr).Once()

	authToken, err := sv.RegisterUser(ctx, login, password)
	require.ErrorAs(s.T(), err, &service.ErrAuth{})
	require.Equal(s.T(), "", authToken)
}

func TestService(t *testing.T) {
	suite.Run(t, new(Suit))
}
