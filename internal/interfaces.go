package internal

import (
	"context"
	"html/template"
	"net/http"
)

type API interface {
	Healthcheck(w http.ResponseWriter, r *http.Request)

	// Register регистрация пользователя
	Register(w http.ResponseWriter, r *http.Request)

	// Login аутентификация пользователя
	Login(w http.ResponseWriter, r *http.Request)

	// PostOrder загрузка пользователем номера заказа для расчёта
	PostOrder(w http.ResponseWriter, r *http.Request)

	// GetOrders получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
	GetOrders(w http.ResponseWriter, r *http.Request)

	// GetBalance получение текущего баланса счёта баллов лояльности пользователя
	GetBalance(w http.ResponseWriter, r *http.Request)

	// Withdraw запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
	Withdraw(w http.ResponseWriter, r *http.Request)

	// Withdrawals - получение информации о выводе средств с накопительного счёта пользователем
	GetWithdrawals(w http.ResponseWriter, r *http.Request)
}

type Templates interface {
	Get(path string) *template.Template
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type SchemaCreator interface {
	SchemaDefined(ctx context.Context) (bool, error)
	CreateSchema(ctx context.Context) error
}

type Resetter interface {
	Reset() error
}
