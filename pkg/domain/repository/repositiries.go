package repository

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
)

type Order interface {
	Insert(ctx context.Context, user *entity.Order) error
	FindByNumber(ctx context.Context, orderNumber string) (*entity.Order, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]entity.Order, error)
	GetAccrualsSum(ctx context.Context, userID uuid.UUID) (float64, error)
	GetWithdrawnSum(ctx context.Context, userID uuid.UUID) (float64, error)
	Withdraw(ctx context.Context, w *entity.Withdraw) error
	GetUserWithdrawals(ctx context.Context, userID uuid.UUID) ([]entity.Withdraw, error)
	SetOrderStatus(ctx context.Context, order *entity.Order, status entity.OrderStatus) error
	UpdateAttributes(ctx context.Context, order *entity.Order) error
}

type User interface {
	Insert(ctx context.Context, user *entity.User) error
	LoginExists(ctx context.Context, login string) (bool, error)
	FindByLoginAndPassword(ctx context.Context, login string, hashedPassword []byte) (*entity.User, error)
}
