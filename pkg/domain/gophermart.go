package domain

import (
	"context"
	"regexp"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/services/accrual"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/repository"
)

const OrderNumberMaxLength = 65535

var _ Gophermart = (*service)(nil)

type Gophermart interface {
	// Register регистрация пользователя
	Register(context.Context, *entity.RegisterRequest) (*entity.User, error)

	// Login аутентификация пользователя
	Login(context.Context, *entity.LoginRequest) (*entity.User, error)

	// PostOrder загрузка пользователем номера заказа для расчёта
	PostOrder(ctx context.Context, order *entity.Order) error

	// GetOrders получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
	GetOrders(ctx context.Context, userID uuid.UUID) ([]entity.Order, error)

	// GetBalance получение текущего баланса счёта баллов лояльности пользователя
	GetBalance(ctx context.Context, userID uuid.UUID) (*entity.Balance, error)

	// Withdraw запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
	Withdraw(ctx context.Context, w *entity.Withdraw) error

	// GetWithdrawals - получение информации о выводе средств с накопительного счёта пользователем
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]entity.Withdraw, error)
}

type Transactor interface {
	Transaction(ctx context.Context, f func(ctx context.Context) error) error
}

type service struct {
	cfg           *config.Config
	trx           Transactor
	hasher        *hasher
	orderNumRegex *regexp.Regexp
	accrual       accrual.Service

	orderRepo repository.Order
	userRepo  repository.User
}

func NewGophermart(
	cfg *config.Config,
	trx Transactor,
	accrual accrual.Service,
	orderRepo repository.Order,
	userRepo repository.User,
) Gophermart {
	return &service{
		cfg:           cfg,
		trx:           trx,
		accrual:       accrual,
		hasher:        &hasher{cfg: &cfg.Auth},
		orderNumRegex: regexp.MustCompile(`^\s*\d+\s*$`),
		orderRepo:     orderRepo,
		userRepo:      userRepo,
	}
}

func (s *service) Register(ctx context.Context, req *entity.RegisterRequest) (*entity.User, error) {
	exists, err := s.userRepo.LoginExists(ctx, req.Login)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrLoginExists
	}

	id, err := uuid.NewV6()
	if err != nil {
		return nil, err
	}
	hashedPwd := s.hasher.GenerateSHA([]byte(req.Password))
	user := &entity.User{
		ID:          id,
		Login:       req.Login,
		PasswordSHA: hashedPwd,
	}
	err = s.userRepo.Insert(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) Login(ctx context.Context, req *entity.LoginRequest) (*entity.User, error) {
	hashedPwd := s.hasher.GenerateSHA([]byte(req.Password))
	user, err := s.userRepo.FindByLoginAndPassword(ctx, req.Login, hashedPwd)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) PostOrder(ctx context.Context, order *entity.Order) error {
	if ok, err := s.CheckOrderNumber(order.Number); err != nil {
		return err
	} else if !ok {
		return ErrBadOrderNumber
	}

	order.Status = entity.OrderStatusNew
	err := s.orderRepo.Insert(ctx, order)
	if err != nil {
		return err
	}

	if err := s.accrual.Send(ctx, &accrual.AccrualEvent{
		Order: order,
	}); err != nil {
		return err
	}

	return nil
}

func (s *service) GetOrders(ctx context.Context, userID uuid.UUID) ([]entity.Order, error) {
	return s.orderRepo.GetUserOrders(ctx, userID)
}

func (s *service) GetBalance(ctx context.Context, userID uuid.UUID) (*entity.Balance, error) {
	var balance entity.Balance
	if err := s.trx.Transaction(ctx, func(ctx context.Context) error {
		var err error
		balance.Current, err = s.orderRepo.GetAccrualsSum(ctx, userID)
		if err != nil {
			return err
		}
		balance.Withdrawn, err = s.orderRepo.GetWithdrawnSum(ctx, userID)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &balance, nil
}

func (s *service) Withdraw(ctx context.Context, w *entity.Withdraw) error {
	if ok, err := s.CheckOrderNumber(w.OrderNumber); err != nil {
		return err
	} else if !ok {
		return ErrBadOrderNumber
	}

	if err := s.trx.Transaction(ctx, func(ctx context.Context) error {
		sum, err := s.orderRepo.GetAccrualsSum(ctx, w.UserID)
		if err != nil {
			return err
		}
		if sum < w.Value {
			return ErrNotEnoughAccruals
		}
		if err := s.orderRepo.Withdraw(ctx, w); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *service) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]entity.Withdraw, error) {
	return s.orderRepo.GetUserWithdrawals(ctx, userID)
}

func (s *service) CheckOrderNumber(number string) (bool, error) {
	if len(number) > OrderNumberMaxLength {
		return false, ErrOrderNumberTooLong
	}
	if !s.orderNumRegex.MatchString(number) {
		return false, ErrBadOrderNumber
	}

	return s.CheckLuhn(number), nil
}

func (s *service) CheckLuhn(number string) bool {
	sum := 0
	nDigits := len(number)
	parity := nDigits % 2
	var digit int
	var err error
	for i := 0; i < nDigits; i++ {
		digit, err = strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}
		if i%2 == parity {
			digit *= 2
		}
		if digit > 9 {
			digit -= 9
			sum += digit
		}
	}

	return (sum % 10) == 0
}
