package entity

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// OrderStatusNew заказ загружен в систему, но не попал в обработку
	OrderStatusNew = OrderStatus("NEW")
	// OrderStatusProcessing вознаграждение за заказ рассчитывается
	OrderStatusProcessing = OrderStatus("PROCESSING")
	// OrderStatusInvalid система расчёта вознаграждений отказала в расчёте
	OrderStatusInvalid = OrderStatus("INVALID")
	// OrderStatusProcessed Данные по заказу проверены и информация о расчёте успешно получена
	OrderStatusProcessed = OrderStatus("PROCESSED")
)

type OrderStatus string

type JwtClaims struct {
	jwt.RegisteredClaims

	UserID uuid.UUID
}

type User struct {
	ID          uuid.UUID `db:"id"`
	Login       string    `db:"login"`
	PasswordSHA []byte    `db:"password_sha"`
}

type Order struct {
	ID        uuid.UUID   `db:"id"`
	UserID    uuid.UUID   `db:"user_id"`
	Number    string      `db:"number"`
	Status    OrderStatus `db:"status"`
	CreatedAt time.Time   `db:"created_at"`
	Accrual   *float64    `db:"accrual"`
}

type Balance struct {
	Current   float64 `db:"current" json:"current"`
	Withdrawn float64 `db:"withdrawn" json:"withdrawn"`
}

type Withdraw struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	OrderNumber string    `db:"order_number"`
	Value       float64   `db:"value"`
	CreatedAt   time.Time `db:"created_at"`
}
