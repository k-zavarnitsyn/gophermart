package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/k-zavarnitsyn/gophermart/internal"
	"github.com/k-zavarnitsyn/gophermart/internal/container"
	"github.com/k-zavarnitsyn/gophermart/internal/middleware"
)

type Router struct {
	*chi.Mux

	cnt *container.Container
}

func NewRouter(cnt *container.Container) *Router {
	return &Router{
		Mux: chi.NewRouter(),
		cnt: cnt,
	}
}

func (r *Router) InitRoutes(a internal.API, withMiddlewares bool) {
	r.Get("/ping", a.Healthcheck)
	r.Post("/api/user/register", a.Register)
	r.Post("/api/user/login", a.Login)

	r.Group(func(router chi.Router) {
		if withMiddlewares {
			authMiddleware := middleware.NewAuth(r.cnt.Auth())
			router.Use(authMiddleware.WithAuthentication)
		}

		router.Post("/api/user/orders", a.PostOrder)
		router.Get("/api/user/orders", a.GetOrders)
		router.Get("/api/user/balance", a.GetBalance)
		router.Post("/api/user/balance/withdraw", a.Withdraw)
		router.Get("/api/user/withdrawals", a.GetWithdrawals)
	})
}
