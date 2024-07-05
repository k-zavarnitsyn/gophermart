package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/repository"
)

const (
	OrderNumberUniqueContraint = "order_number_uindex"
)

var _ repository.Order = &OrderRepo{}

type OrderRepo struct {
	db *Pool
}

func NewOrderRepository(db *Pool) repository.Order {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Insert(ctx context.Context, order *entity.Order) error {
	if order.ID.IsNil() {
		var err error
		if order.ID, err = uuid.NewV6(); err != nil {
			return err
		}
	}

	sql := `
		INSERT INTO "order" (id, user_id, number, status, accrual)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, sql, order.ID, order.UserID, order.Number, order.Status, order.Accrual)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == OrderNumberUniqueContraint {
			existingOrder, err2 := r.FindByNumber(ctx, order.Number)
			if err2 != nil {
				return fmt.Errorf("error searching order on number unique violation: %w", errors.Join(err, err2))
			}
			if existingOrder.UserID == order.UserID {
				return domain.ErrOrderCreatedByCurrentUser
			}
			return domain.ErrOrderCreatedByOtherUser
		}
		return err
	}

	return nil
}

func (r *OrderRepo) FindByNumber(ctx context.Context, orderNumber string) (*entity.Order, error) {
	var value entity.Order
	sql := `SELECT * FROM "order" WHERE number = $1;`
	err := pgxscan.Get(ctx, r.db, &value, sql, orderNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &value, err
}

func (r *OrderRepo) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]entity.Order, error) {
	var values []entity.Order
	sql := `SELECT * FROM "order" WHERE user_id = $1;`
	err := pgxscan.Select(ctx, r.db, &values, sql, userID)
	if err != nil {
		return nil, err
	}

	return values, err
}

func (r *OrderRepo) GetAccrualsSum(ctx context.Context, userID uuid.UUID) (float64, error) {
	var value *float64
	sql := `SELECT sum(accrual) FROM "order" WHERE user_id = $1 AND status = $2;`
	err := pgxscan.Get(ctx, r.db, &value, sql, userID, entity.OrderStatusProcessed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrNotFound
		}
		return 0, err
	}
	if value == nil {
		return 0, nil
	}

	return *value, err
}

func (r *OrderRepo) GetWithdrawnSum(ctx context.Context, userID uuid.UUID) (float64, error) {
	var value *float64
	sql := `SELECT sum(value) FROM withdrawn WHERE user_id = $1;`
	err := pgxscan.Get(ctx, r.db, &value, sql, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrNotFound
		}
		return 0, err
	}
	if value == nil {
		return 0, nil
	}

	return *value, err
}

func (r *OrderRepo) Withdraw(ctx context.Context, w *entity.Withdraw) error {
	if w.ID.IsNil() {
		var err error
		if w.ID, err = uuid.NewV6(); err != nil {
			return err
		}
	}

	sql := `
		INSERT INTO withdrawn (id, user_id, order_number, value)
		VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, sql, w.ID, w.UserID, w.OrderNumber, w.Value)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrderRepo) GetUserWithdrawals(ctx context.Context, userID uuid.UUID) ([]entity.Withdraw, error) {
	var values []entity.Withdraw
	sql := `SELECT * FROM withdrawn WHERE user_id = $1;`
	err := pgxscan.Select(ctx, r.db, &values, sql, userID)
	if err != nil {
		return nil, err
	}

	return values, err
}

func (r *OrderRepo) SetOrderStatus(ctx context.Context, orderNumber string, status entity.OrderStatus) error {
	sql := `
		UPDATE "order" SET status = $1
		WHERE number = $2`
	_, err := r.db.Exec(ctx, sql, status, orderNumber)
	if err != nil {
		return err
	}

	return nil
}
func (r *OrderRepo) UpdateAttributes(ctx context.Context, order *entity.Order) error {
	if order.ID.IsNil() {
		return domain.NewError("unable to update order: ID not set")
	}

	sql := `
		UPDATE "order" SET (status, accrual) = ($2, $3)
		WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, order.ID, order.Status, order.Accrual)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrderRepo) GetOrdersByStatuses(ctx context.Context, statuses []string, exceptNumbers []string, limit int) ([]entity.Order, error) {
	var values []entity.Order
	var err error
	if len(statuses) == 0 {
		return nil, fmt.Errorf("unable to get orders: no statuses provided")
	}
	if len(exceptNumbers) == 0 {
		sql := `SELECT * FROM "order" WHERE status = ANY($1) LIMIT $2;`
		err = pgxscan.Select(ctx, r.db, &values, sql, statuses, limit)
	} else {
		sql := `SELECT * FROM "order" WHERE status = ANY($1) and number <> ANY ($2) LIMIT $3;`
		err = pgxscan.Select(ctx, r.db, &values, sql, statuses, exceptNumbers, limit)
	}
	if err != nil {
		return nil, err
	}

	return values, err
}
