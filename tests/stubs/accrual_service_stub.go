package stubs

import (
	"context"

	"github.com/k-zavarnitsyn/gophermart/internal/services/accrual"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
)

var _ accrual.Service = (*AccrualServiceStub)(nil)

type AccrualServiceStub struct {
	orders []*entity.Order
}

func (a *AccrualServiceStub) Send(ctx context.Context, order *entity.Order) error {
	a.orders = append(a.orders, order)
	return nil
}

func (a *AccrualServiceStub) Orders() []*entity.Order {
	return a.orders
}
