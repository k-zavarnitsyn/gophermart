package container

import (
	"github.com/k-zavarnitsyn/gophermart/internal"
	"github.com/k-zavarnitsyn/gophermart/internal/pg"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/repository"
)

func (c *Container) getUtilityRepo() *pg.UtilityRepository {
	if c.utilityRepo == nil {
		c.utilityRepo = pg.NewUtilityRepository(c.DB())
	}

	return c.utilityRepo
}

func (c *Container) Pinger() internal.Pinger {
	return c.getUtilityRepo()
}

func (c *Container) SchemaCreator() internal.SchemaCreator {
	return c.getUtilityRepo()
}

func (c *Container) Resetter() internal.Resetter {
	return c.getUtilityRepo()
}

func (c *Container) OrderRepo() repository.Order {
	if c.orderRepo == nil {
		c.orderRepo = pg.NewOrderRepository(c.DB())
	}

	return c.orderRepo
}

func (c *Container) UserRepo() repository.User {
	if c.userRepo == nil {
		c.userRepo = pg.NewUserRepository(c.DB())
	}

	return c.userRepo
}
