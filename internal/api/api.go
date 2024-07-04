package api

import (
	"github.com/k-zavarnitsyn/gophermart/internal"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/services/auth"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
)

var _ internal.API = (*gophermartServer)(nil)

type gophermartServer struct {
	cfg        *config.Config
	auth       *auth.Service
	gophermart domain.Gophermart

	dbPinger internal.Pinger
}

func New(
	cfg *config.Config,
	authService *auth.Service,
	service domain.Gophermart,
	dbPinger internal.Pinger,
) internal.API {
	return &gophermartServer{
		cfg:        cfg,
		auth:       authService,
		gophermart: service,
		dbPinger:   dbPinger,
	}
}
