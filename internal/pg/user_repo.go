package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/repository"
)

type UserRepo struct {
	db *Pool
}

func NewUserRepository(db *Pool) repository.User {
	return &UserRepo{db: db}
}

func (r *UserRepo) Insert(ctx context.Context, user *entity.User) error {
	sql := `
		INSERT INTO "user" (id, login, password_sha)
		VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, sql, user.ID, user.Login, user.PasswordSHA)

	return err
}

func (r *UserRepo) LoginExists(ctx context.Context, login string) (bool, error) {
	var value int
	sql := `select 1 from "user" where login = $1;`
	err := pgxscan.Get(ctx, r.db, &value, sql, login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *UserRepo) FindByLoginAndPassword(ctx context.Context, login string, hashedPassword []byte) (*entity.User, error) {
	var value entity.User
	sql := `select * from "user" where login = $1 and password_sha = $2;`
	err := pgxscan.Get(ctx, r.db, &value, sql, login, hashedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("login and password: %w", domain.ErrNotFound)
		}
		return nil, err
	}

	return &value, nil
}
