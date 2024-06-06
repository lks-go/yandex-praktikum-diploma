package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/lks-go/yandex-praktikum-diploma/internal/service"
	"github.com/lks-go/yandex-praktikum-diploma/internal/service/mocks"
)

var ErrAny = errors.New("any unexpected error")

type Suite struct {
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

func (s *Suite) SetupTest() {
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

func (s *Suite) TestService_RegisterUser_Positive() {
	ctx := context.Background()
	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6InRlc3RfdXNlciIsImlhdCI6MTUxNjIzOTAyMn0.YTH8MZcIu-j5Fw7fr2zi4KB52c1x0P1d2XlUJ7fak1o"

	deps := service.Deps{
		UserStorage:  s.userStorage,
		TokenBuilder: s.tokenBuilder,
	}
	sv := service.New(&s.serviceConfig, &deps)

	login := "test_user"
	password := "test_password"

	s.userStorage.On("AddUser", ctx, login, sv.HashPassword(password)).
		Return("user-id", nil).Once()

	s.tokenBuilder.On("BuildNewToken", login).Return(testToken, nil).Once()

	authToken, err := sv.RegisterUser(ctx, login, password)
	require.NoError(s.T(), err)
	require.Equal(s.T(), testToken, authToken)
}

func (s *Suite) TestService_RegisterUser_NegativeAddUser() {
	ctx := context.Background()

	deps := service.Deps{
		UserStorage: s.userStorage,
	}
	sv := service.New(&s.serviceConfig, &deps)

	login := "test_user"
	password := "test_password"

	s.userStorage.On("AddUser", ctx, login, sv.HashPassword(password)).
		Return("", ErrAny).Once()

	authToken, err := sv.RegisterUser(ctx, login, password)
	require.Error(s.T(), err)
	require.Equal(s.T(), "", authToken)
}

func (s *Suite) TestService_RegisterUser_NegativeBuildToken() {
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

	tokenErr := fmt.Errorf("failed to build token: %w", service.ErrAuth{Err: ErrAny})
	s.tokenBuilder.On("BuildNewToken", login).
		Return("", tokenErr).Once()

	authToken, err := sv.RegisterUser(ctx, login, password)
	require.ErrorAs(s.T(), err, &service.ErrAuth{})
	require.Equal(s.T(), "", authToken)
}

func (s *Suite) TestService_AuthUser_Positive() {
	ctx := context.Background()
	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6InRlc3RfdXNlcl8yIiwiaWF0IjoxNTE2MjM5MDIyfQ.p7EiLJUzXtKQ3zdlpfQfbsWZzVreLCsDb7sb93pPKu0"

	deps := service.Deps{
		UserStorage:  s.userStorage,
		TokenBuilder: s.tokenBuilder,
	}
	sv := service.New(&s.serviceConfig, &deps)

	password := "test_password_2"
	u := service.User{
		ID:           uuid.NewString(),
		Login:        "test_user_2",
		PasswordHash: sv.HashPassword(password),
	}

	s.userStorage.On("UserByLogin", ctx, u.Login).Return(&u, nil).Once()
	s.tokenBuilder.On("BuildNewToken", u.Login).Return(testToken, nil).Once()

	authToken, err := sv.AuthUser(ctx, u.Login, password)
	require.NoError(s.T(), err)
	require.Equal(s.T(), testToken, authToken)
}

func (s *Suite) TestService_AuthUser_NegativeUserByLogin() {
	ctx := context.Background()

	deps := service.Deps{
		UserStorage:  s.userStorage,
		TokenBuilder: s.tokenBuilder,
	}
	sv := service.New(&s.serviceConfig, &deps)

	login := "test_user_2"
	password := "test_password_2"

	s.userStorage.On("UserByLogin", ctx, login).Return(nil, ErrAny).Once()

	authToken, err := sv.AuthUser(ctx, login, password)
	require.Error(s.T(), err)
	require.Equal(s.T(), "", authToken)
}

func (s *Suite) TestService_AuthUser_NegativePasswordNotMatch() {
	ctx := context.Background()

	deps := service.Deps{
		UserStorage:  s.userStorage,
		TokenBuilder: s.tokenBuilder,
	}
	sv := service.New(&s.serviceConfig, &deps)

	password := "test_password_2"
	passwordWrong := "test_password_wrong"
	u := service.User{
		ID:           uuid.NewString(),
		Login:        "test_user_2",
		PasswordHash: sv.HashPassword(password),
	}

	s.userStorage.On("UserByLogin", ctx, u.Login).Return(&u, nil).Once()

	authToken, err := sv.AuthUser(ctx, u.Login, passwordWrong)
	require.ErrorIs(s.T(), err, service.ErrUsersPasswordNotMatch)
	require.Equal(s.T(), "", authToken)
}

func (s *Suite) TestService_AuthUser_NegativeBuildToken() {
	ctx := context.Background()

	deps := service.Deps{
		UserStorage:  s.userStorage,
		TokenBuilder: s.tokenBuilder,
	}
	sv := service.New(&s.serviceConfig, &deps)

	password := "test_password_2"
	u := service.User{
		ID:           uuid.NewString(),
		Login:        "test_user_2",
		PasswordHash: sv.HashPassword(password),
	}

	s.userStorage.On("UserByLogin", ctx, u.Login).Return(&u, nil).Once()

	tokenErr := fmt.Errorf("failed to build token: %w", service.ErrAuth{Err: ErrAny})
	s.tokenBuilder.On("BuildNewToken", u.Login).Return("", tokenErr).Once()

	authToken, err := sv.AuthUser(ctx, u.Login, password)
	require.ErrorAs(s.T(), err, &service.ErrAuth{})
	require.Equal(s.T(), "", authToken)

}

func (s *Suite) TestService_SaveOrder_PositiveAddOrder() {
	ctx := context.Background()
	userID := uuid.NewString()
	newOrderID := uuid.NewString()
	login := "user"
	orderNumber := "123"

	deps := service.Deps{
		UserStorage:           s.userStorage,
		OrderStorage:          s.orderStorage,
		OrderProcessPublisher: s.orderProcessPublisher,
	}
	sv := service.New(&s.serviceConfig, &deps)

	s.userStorage.On("UserByLogin", ctx, login).Return(&service.User{ID: userID, Login: login}, nil).Once()
	s.orderStorage.On("OrderByNumber", ctx, orderNumber).Return(nil, nil).Once()

	newOrder := service.Order{UserID: userID, Number: orderNumber, Status: service.OrderStatusNew}
	s.orderStorage.On("AddOrder", ctx, &newOrder).Return(newOrderID, nil).Once()

	event := service.OrderEvent{UserID: userID, OrderID: newOrderID, OrderNumber: orderNumber}
	s.orderProcessPublisher.On("Publish", ctx, event).Once()

	err := sv.SaveOrder(ctx, login, orderNumber)
	require.NoError(s.T(), err)

	// waiting for calling Publish in goroutine
	time.Sleep(time.Millisecond * 5)
}

func (s *Suite) TestService_SaveOrder_PositiveAlreadyExists() {
	ctx := context.Background()
	userID := uuid.NewString()
	login := "user"
	orderNumber := "123"

	deps := service.Deps{
		UserStorage:           s.userStorage,
		OrderStorage:          s.orderStorage,
		OrderProcessPublisher: s.orderProcessPublisher,
	}
	sv := service.New(&s.serviceConfig, &deps)

	s.userStorage.On("UserByLogin", ctx, login).Return(&service.User{ID: userID, Login: login}, nil).Once()
	s.orderStorage.On("OrderByNumber", ctx, orderNumber).Return(&service.Order{UserID: userID}, nil).Once()

	err := sv.SaveOrder(ctx, login, orderNumber)
	require.ErrorIs(s.T(), err, service.ErrAlreadyExists)
}

func (s *Suite) TestService_SaveOrder_NegativeConflict() {
	ctx := context.Background()
	userID := uuid.NewString()
	login := "user"
	orderNumber := "123"

	deps := service.Deps{
		UserStorage:           s.userStorage,
		OrderStorage:          s.orderStorage,
		OrderProcessPublisher: s.orderProcessPublisher,
	}
	sv := service.New(&s.serviceConfig, &deps)

	s.userStorage.On("UserByLogin", ctx, login).Return(&service.User{ID: userID, Login: login}, nil).Once()
	s.orderStorage.On("OrderByNumber", ctx, orderNumber).
		Return(&service.Order{UserID: uuid.NewString()}, nil).Once()

	err := sv.SaveOrder(ctx, login, orderNumber)
	require.ErrorIs(s.T(), err, service.ErrOrderConflict)
}

func TestService(t *testing.T) {
	suite.Run(t, new(Suite))
}
