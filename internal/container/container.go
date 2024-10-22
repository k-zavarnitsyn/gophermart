package container

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/pg"
	"github.com/k-zavarnitsyn/gophermart/internal/services/accrual"
	"github.com/k-zavarnitsyn/gophermart/internal/services/auth"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/repository"
	log "github.com/sirupsen/logrus"
)

type Container struct {
	cfg *config.Config
	db  *pg.Pool
	trx pg.Transactor

	auth              *auth.Service
	accrualService    accrual.Service
	gophermartService domain.Gophermart

	utilityRepo *pg.UtilityRepository
	orderRepo   repository.Order
	userRepo    repository.User
}

func New(cfg *config.Config) *Container {
	return &Container{
		cfg: cfg,
	}
}

func (c *Container) Shutdown(ctx context.Context) error {
	return nil
}

func (c *Container) DB() *pg.Pool {
	if c.db == nil {
		conf, err := pgxpool.ParseConfig(c.cfg.Server.DatabaseURI)
		if err != nil {
			log.WithError(err).Fatal("failed to parse postgres config")
		}

		ctx := context.Background()
		pool, err := pgxpool.NewWithConfig(ctx, conf)
		if err != nil {
			log.WithError(err).Fatal("failed to connect to postgres")
		}

		c.db = pg.NewPool(pool)
	}

	return c.db
}

func (c *Container) Transactor() pg.Transactor {
	if c.trx == nil {
		c.trx = pg.NewTransactor(c.DB())
	}

	return c.trx
}

func (c *Container) Auth() *auth.Service {
	if c.auth == nil {
		c.auth = auth.New(&c.cfg.Auth)
	}

	return c.auth
}

func (c *Container) Gophermart() domain.Gophermart {
	if c.gophermartService == nil {
		c.gophermartService = domain.NewGophermart(c.cfg, c.Transactor(), c.AccrualService(), c.OrderRepo(), c.UserRepo())
	}

	return c.gophermartService
}

func (c *Container) AccrualService() accrual.Service {
	if c.accrualService == nil {
		c.accrualService = accrual.NewService(&c.cfg.Accrual, c.OrderRepo())
	}

	return c.accrualService
}

func (c *Container) SetAccrualService(s accrual.Service) *Container {
	c.accrualService = s

	return c
}
