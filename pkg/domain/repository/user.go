package repository

import (
	"context"

	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
)

type User interface {
	Insert(ctx context.Context, user *entity.User) error
	LoginExists(ctx context.Context, login string) (bool, error)
	FindByLoginAndPassword(ctx context.Context, login string, hashedPassword []byte) (*entity.User, error)
}
