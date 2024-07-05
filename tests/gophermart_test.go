package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/k-zavarnitsyn/gophermart/internal"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/container"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
	"github.com/k-zavarnitsyn/gophermart/tests/stubs"
	"github.com/k-zavarnitsyn/gophermart/tests/testutils"
	"github.com/stretchr/testify/suite"
)

type GophermartTestSuite struct {
	suite.Suite

	cfg  *config.Config
	cnt  *container.Container
	api  internal.API
	user *entity.User

	userCount int
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, new(GophermartTestSuite))
}

func (s *GophermartTestSuite) SetupSuite() {
	s.cfg = testutils.GetConfig("../" + config.DefaultDir)
	s.cnt = container.New(s.cfg)
	s.cnt.SetAccrualService(&stubs.AccrualServiceStub{})

	s.Require().NoError(s.cnt.Resetter().Reset())

	s.user = s.NewUser()
}

func (s *GophermartTestSuite) NewUser() *entity.User {
	s.userCount++
	u, err := s.cnt.Gophermart().Register(context.Background(), &entity.RegisterRequest{
		Login:    "test" + strconv.Itoa(s.userCount),
		Password: "test",
	})
	s.Require().NoError(err)
	s.Require().NotNil(u)

	return u
}

func (s *GophermartTestSuite) TestLoginSuccess() {
	u, err := s.cnt.Gophermart().Login(context.Background(), &entity.LoginRequest{
		Login:    s.user.Login,
		Password: "test",
	})
	s.Require().NoError(err)
	s.Require().NotNil(u)
}

func (s *GophermartTestSuite) TestLoginFailed() {
	u, err := s.cnt.Gophermart().Login(context.Background(), &entity.LoginRequest{
		Login:    s.user.Login,
		Password: "wrong",
	})
	s.Require().ErrorIs(err, domain.ErrNotFound)
	s.Require().Nil(u)
}

func (s *GophermartTestSuite) TestPostOrder() {
	err := s.cnt.Gophermart().PostOrder(context.Background(), &entity.Order{
		UserID:    s.user.ID,
		Number:    "3413042486",
		Status:    entity.OrderStatusNew,
		CreatedAt: time.Now(),
	})
	s.Require().NoError(err)
	accService, ok := s.cnt.AccrualService().(*stubs.AccrualServiceStub)
	s.Require().True(ok)
	s.Require().True(utils.ContainsWhere(accService.Orders(), func(o *entity.Order) bool {
		return o.Number == "3413042486"
	}))

	err = s.cnt.Gophermart().PostOrder(context.Background(), &entity.Order{
		UserID:    s.user.ID,
		Number:    "3413042486",
		Status:    entity.OrderStatusNew,
		CreatedAt: time.Now(),
	})
	s.Require().ErrorIs(err, domain.ErrOrderCreatedByCurrentUser)
}

func (s *GophermartTestSuite) TestPostOrderBadNumber() {
	err := s.cnt.Gophermart().PostOrder(context.Background(), &entity.Order{
		UserID:    s.user.ID,
		Number:    "3413042466",
		Status:    entity.OrderStatusNew,
		CreatedAt: time.Now(),
	})
	s.Require().ErrorIs(err, domain.ErrBadOrderNumber)
}
func (s *GophermartTestSuite) TestGetOrders() {
	u := s.NewUser()
	order := &entity.Order{
		ID:     uuid.Must(uuid.NewV6()),
		UserID: u.ID,
		Number: "5798116405",
		Status: entity.OrderStatusNew,
	}
	err := s.cnt.Gophermart().PostOrder(context.Background(), order)
	s.Require().NoError(err)

	orders, err := s.cnt.Gophermart().GetOrders(context.Background(), u.ID)
	s.Require().NoError(err)
	s.Require().True(utils.ContainsWhere(orders, func(e entity.Order) bool {
		return e.Number == order.Number
	}))
}

func (s *GophermartTestSuite) TestBalance() {
	u := s.NewUser()
	orders := []entity.Order{
		{
			ID:      uuid.Must(uuid.NewV6()),
			UserID:  u.ID,
			Number:  "9155976989",
			Status:  entity.OrderStatusNew,
			Accrual: utils.ToPointer(80.00),
		},
		{
			ID:      uuid.Must(uuid.NewV6()),
			UserID:  u.ID,
			Number:  "1587579366",
			Status:  entity.OrderStatusProcessing,
			Accrual: utils.ToPointer(40.00),
		},
		{
			ID:      uuid.Must(uuid.NewV6()),
			UserID:  u.ID,
			Number:  "3203697697",
			Status:  entity.OrderStatusProcessed,
			Accrual: utils.ToPointer(20.00),
		},
		{
			ID:      uuid.Must(uuid.NewV6()),
			UserID:  u.ID,
			Number:  "6409723027",
			Status:  entity.OrderStatusProcessed,
			Accrual: utils.ToPointer(10.05),
		},
	}
	for _, order := range orders {
		err := s.cnt.OrderRepo().Insert(context.Background(), &order)
		s.Require().NoError(err)
	}

	b, err := s.cnt.Gophermart().GetBalance(context.Background(), u.ID)
	s.Require().NoError(err)
	s.Require().Equal(30.05, b.Current)
}

func (s *GophermartTestSuite) TestWithdraw() {
	u := s.NewUser()
	orders := []entity.Order{
		{
			ID:      uuid.Must(uuid.NewV6()),
			UserID:  u.ID,
			Number:  "0928953488",
			Status:  entity.OrderStatusProcessed,
			Accrual: utils.ToPointer(80.00),
		},
		{
			ID:      uuid.Must(uuid.NewV6()),
			UserID:  u.ID,
			Number:  "9325279751",
			Status:  entity.OrderStatusProcessed,
			Accrual: utils.ToPointer(40.00),
		},
	}
	for _, order := range orders {
		err := s.cnt.OrderRepo().Insert(context.Background(), &order)
		s.Require().NoError(err)
	}

	err := s.cnt.Gophermart().Withdraw(context.Background(), &entity.Withdraw{
		UserID:      u.ID,
		OrderNumber: "9325279751",
		Value:       10.00,
	})
	s.Require().NoError(err)
	err = s.cnt.Gophermart().Withdraw(context.Background(), &entity.Withdraw{
		UserID:      u.ID,
		OrderNumber: "9325279751",
		Value:       3.00,
	})
	s.Require().NoError(err)
	err = s.cnt.Gophermart().Withdraw(context.Background(), &entity.Withdraw{
		UserID:      u.ID,
		OrderNumber: "0928953488",
		Value:       153.00,
	})
	s.Require().ErrorIs(err, domain.ErrNotEnoughAccruals)

	b, err := s.cnt.Gophermart().GetBalance(context.Background(), u.ID)
	s.Require().NoError(err)
	s.Require().Equal(107.00, b.Current)
	s.Require().Equal(13.00, b.Withdrawn)
}
