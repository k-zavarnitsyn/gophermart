package pg

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
)

type Pool RetriablePool

type Querier interface {
	pgxscan.Querier

	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type RetriablePool struct {
	*pgxpool.Pool
}

type retriablePostgresErr struct {
	err error
}

func NewPool(pool *pgxpool.Pool) *Pool {
	return &Pool{pool}
}

func (e *retriablePostgresErr) Error() string {
	return e.err.Error()
}

func (e *retriablePostgresErr) IsRetriable() bool {
	var pgErr *pgconn.PgError

	return pgconn.SafeToRetry(e.err) || (errors.As(e.err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code))
}

func (p *Pool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	var tag pgconn.CommandTag
	err := utils.Retry(func() error {
		var err error
		tag, err = p.Pool.Exec(ctx, sql, arguments...)
		if err != nil {
			return &retriablePostgresErr{err}
		}

		return nil
	})

	return tag, err
}

func (p *Pool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	var rows pgx.Rows
	err := utils.Retry(func() error {
		var err error
		rows, err = p.Pool.Query(ctx, sql, args...)
		if err != nil {
			return &retriablePostgresErr{err}
		}

		return nil
	})

	return rows, err
}

// Transaction запускает все запросы функции внутри транзакции (необходимо использовать параметр tx, иначе пул самоисчерпается).
// Возврат ошибки приводит к откату.
func (p *Pool) Transaction(ctx context.Context, f func(ctx context.Context, tx pgx.Tx) error) (err error) {
	tx := TxFromContext(ctx)
	if tx == nil {
		tx, err = p.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		ctx = TxToContext(ctx, tx)

		defer func() {
			err = commitOrRollbackPGX(ctx, tx, err, recover())
		}()
	}

	return f(ctx, tx)
}
