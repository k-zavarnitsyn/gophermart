package accrual

import (
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
)

type Event interface {
	GetOrderNumber() string
}

type AccrualEvent struct {
	Order *entity.Order
}

type AccrualResponse struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual"`
}
