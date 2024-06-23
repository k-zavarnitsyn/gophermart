package domain

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/repository"
)

var _ Gophermart = (*service)(nil)

type Gophermart interface {
	// Register регистрация пользователя
	Register(context.Context, *entity.RegisterRequest) (*entity.User, error)

	// Login аутентификация пользователя
	Login(context.Context, *entity.LoginRequest) (*entity.User, error)

	// PostOrders загрузка пользователем номера заказа для расчёта
	PostOrders(ctx context.Context, userID uuid.UUID) error

	// GetOrders получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
	GetOrders(ctx context.Context, userID uuid.UUID) error

	// GetBalance получение текущего баланса счёта баллов лояльности пользователя
	GetBalance(ctx context.Context, userID uuid.UUID) error

	// Withdraw запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
	Withdraw(ctx context.Context, userID uuid.UUID, r *entity.WithdrawRequest) error

	// GetWithdrawals - получение информации о выводе средств с накопительного счёта пользователем
	GetWithdrawals(ctx context.Context, userID uuid.UUID) error
}

type service struct {
	cfg    *config.Config
	hasher *hasher

	orderRepo repository.Order
	userRepo  repository.User
}

func NewGophermart(cfg *config.Config, orderRepo repository.Order, userRepo repository.User) Gophermart {
	return &service{
		cfg:       cfg,
		hasher:    &hasher{cfg: &cfg.Auth},
		orderRepo: orderRepo,
		userRepo:  userRepo,
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

func (s *service) PostOrders(ctx context.Context, userID uuid.UUID) error {
	// TODO implement me
	panic("implement me")
}

func (s *service) GetOrders(ctx context.Context, userID uuid.UUID) error {
	// TODO implement me
	panic("implement me")
}

func (s *service) GetBalance(ctx context.Context, userID uuid.UUID) error {
	// TODO implement me
	panic("implement me")
}

func (s *service) Withdraw(ctx context.Context, userID uuid.UUID, req *entity.WithdrawRequest) error {
	// TODO implement me
	panic("implement me")
}

func (s *service) GetWithdrawals(ctx context.Context, userID uuid.UUID) error {
	// TODO implement me
	panic("implement me")
}
