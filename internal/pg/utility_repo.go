package pg

import (
	"context"
	"fmt"

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
		create table if not exists public.counters
		(
			id    uuid primary key,
			name  varchar          not null,
			value bigint default 0 not null
		);
		create unique index if not exists counters_name_index on public.counters (name);
		
		create table if not exists public.gauges
		(
			id    uuid primary key,
			name  varchar          not null,
			value double precision default 0 not null
		);
		create index if not exists gauges_name_index on public.gauges (name);`
	_, err := r.db.Exec(ctx, sql)

	return err
}

func (r *UtilityRepository) SchemaDefined(ctx context.Context) (bool, error) {
	if exists, err := r.TableExists(ctx, "counters"); err != nil || !exists {
		return exists, err
	}
	if exists, err := r.TableExists(ctx, "gauges"); err != nil || !exists {
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
	if err := r.Truncate(context.Background(), "counters"); err != nil {
		return err
	}
	if err := r.Truncate(context.Background(), "gauges"); err != nil {
		return err
	}

	return nil
}

func (r *UtilityRepository) Truncate(ctx context.Context, table string) error {
	sql := fmt.Sprintf(`truncate table %q;`, table)
	_, err := r.db.Exec(ctx, sql)

	return err
}
