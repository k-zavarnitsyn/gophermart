package entity

import (
	"time"
)

type OrderResponse struct {
	Number    string      `json:"number"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"uploaded_at"`
	Accrual   *float64    `json:"accrual,omitempty"`
}

type WithdrawalsResponse struct {
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
