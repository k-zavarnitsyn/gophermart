package container

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/pg"
	"github.com/k-zavarnitsyn/gophermart/internal/services/auth"
	"github.com/k-zavarnitsyn/gophermart/internal/templates"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/repository"
	log "github.com/sirupsen/logrus"
)

type Container struct {
	cfg       *config.Config
	db        *pg.Pool
	templates *templates.Loader

	auth              *auth.Service
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

func (c *Container) Auth() *auth.Service {
	if c.auth == nil {
		c.auth = auth.New(&c.cfg.Auth)
	}

	return c.auth
}

func (c *Container) Templates() *templates.Loader {
	if c.templates == nil {
		t, err := templates.NewTemplateLoader(templates.Path)
		if err != nil {
			panic(err)
		}
		c.templates = t
	}

	return c.templates
}

func (c *Container) Gophermart() domain.Gophermart {
	if c.gophermartService == nil {
		c.gophermartService = domain.NewGophermart(c.cfg, c.OrderRepo(), c.UserRepo())
	}

	return c.gophermartService
}
