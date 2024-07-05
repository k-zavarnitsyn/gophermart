package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
)

type UtilityRepository struct {
	db *Pool
}

func NewUtilityRepository(db *Pool) *UtilityRepository {
	return &UtilityRepository{db: db}
}

func (r *UtilityRepository) Ping(ctx context.Context) error {
	_, err := r.db.Exec(ctx, "SELECT 1")

	return err
}

func (r *UtilityRepository) CreateSchema(ctx context.Context) error {
	sql := `
		create type order_status as enum ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
		create table if not exists "user"
		(
			id uuid not null
				constraint user_pk
					primary key,
			password_sha bytea not null,
			login varchar(64) not null
		);
		
		create table if not exists "order"
		(
			id uuid not null
				constraint order_pk
					primary key,
			number varchar not null,
			user_id uuid not null
				constraint order_user_id_fk
					references "user",
			created_at timestamp default now() not null,
			status order_status default 'NEW'::order_status not null,
			accrual double precision
		);
		
		create index if not exists order_user_id_status_index
			on "order" (user_id, status);
		
		create unique index if not exists order_number_uindex
			on "order" (number);
		
		create table if not exists withdrawn
		(
			id uuid not null
				constraint withdrawn_pk
					primary key,
			user_id uuid not null
				constraint withdrawn_user_id_fk
					references "user",
			value double precision not null,
			created_at timestamp default now() not null,
			order_number varchar not null
		);
		
		create index if not exists withdrawn_user_id_index
			on withdrawn (user_id);`
	_, err := r.db.Exec(ctx, sql)

	return err
}

func (r *UtilityRepository) SchemaDefined(ctx context.Context) (bool, error) {
	if exists, err := r.TableExists(ctx, "order"); err != nil || !exists {
		return exists, err
	}
	if exists, err := r.TableExists(ctx, "user"); err != nil || !exists {
		return exists, err
	}
	if exists, err := r.TableExists(ctx, "withdrawn"); err != nil || !exists {
		return exists, err
	}

	return true, nil
}

func (r *UtilityRepository) TableExists(ctx context.Context, name string) (bool, error) {
	sql := `SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_name = $1
	);`
	exists := false
	err := pgxscan.Get(ctx, r.db, &exists, sql, name)

	return exists, err
}

func (r *UtilityRepository) Reset() error {
	if err := r.Truncate(context.Background(), "withdrawn", "order", "user"); err != nil {
		return err
	}

	return nil
}

func (r *UtilityRepository) Truncate(ctx context.Context, tables ...string) error {
	sql := fmt.Sprintf(`truncate table "%s";`, strings.Join(tables, `","`)) //nolint:gocritic // false positive
	_, err := r.db.Exec(ctx, sql)

	return err
}
