package api

import (
	"github.com/k-zavarnitsyn/gophermart/internal"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/services/auth"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
)

const (
	ContentType         = "Content-Type"
	ContentTypeText     = "text/plain"
	ContentTypeTextHTML = "text/html"
	ContentTypeJSON     = "application/json"
	ContentEncoding     = "Content-Encoding"
	AcceptEncoding      = "Accept-Encoding"
)

var _ internal.API = (*gophermartServer)(nil)

type gophermartServer struct {
	cfg        *config.Config
	auth       *auth.Service
	gophermart domain.Gophermart
	templates  internal.Templates

	dbPinger internal.Pinger
}

func New(
	cfg *config.Config,
	authService *auth.Service,
	service domain.Gophermart,
	tpl internal.Templates,
	dbPinger internal.Pinger,
) internal.API {
	return &gophermartServer{
		cfg:        cfg,
		auth:       authService,
		gophermart: service,
		templates:  tpl,
		dbPinger:   dbPinger,
	}
}
