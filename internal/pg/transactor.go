package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type trxKey struct{}

// Transactor умеет запускать транзакции в сервисах, помещая их в контекст.
// Сервисы при этом ничего не знают про БД, а реализовать транзакцию можно кастомным способом. Но нам достаточно БД.
type Transactor interface {
	Transaction(ctx context.Context, f func(ctx context.Context) error) error
}

type transactor struct {
	db *Pool
}

func NewTransactor(db *Pool) Transactor {
	return &transactor{db: db}
}

func (t *transactor) Transaction(ctx context.Context, f func(ctx context.Context) error) error {
	return t.db.Transaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return f(ctx)
	})
}

func TxFromContext(ctx context.Context) pgx.Tx {
	tx, ok := ctx.Value(&trxKey{}).(pgx.Tx)
	if !ok {
		return nil
	}

	return tx
}

func TxToContext(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, &trxKey{}, tx)
}

func commitOrRollbackPGX(ctx context.Context, tx pgx.Tx, err error, panicErr any) error {
	if panicErr != nil {
		var innerErr error
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			innerErr = fmt.Errorf("panic: %v (unable to rollback: %w)", panicErr, rollbackErr)
		} else {
			innerErr = fmt.Errorf("panic: %v", panicErr)
		}
		if err != nil {
			err = errors.Join(err, innerErr)
		} else {
			err = innerErr
		}

		return err
	}

	if err == nil {
		return tx.Commit(ctx)
	}
	if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}
