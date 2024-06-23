package pg

import (
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/repository"
)

var _ repository.Order = &OrderRepo{}

type OrderRepo struct {
	db *Pool
}

func NewOrderRepository(db *Pool) repository.Order {
	return &OrderRepo{db: db}
}

// func (s *OrderRepo) IncCounter(name string, value int64) error {
// 	id, err := uuid.NewV6()
// 	if err != nil {
// 		return err
// 	}
//
// 	sql := `
// 		INSERT INTO public.counters (id, name, "value")
// 		VALUES ($1, $2, $3)
// 		ON CONFLICT (name) DO UPDATE
// 		  SET "value" = counters."value" + $3;`
// 	_, err = s.db.Exec(context.Background(), sql, id, name, value)
//
// 	return err
// }
//
// func (s *OrderRepo) AppendGauge(name string, value float64) error {
// 	id, err := uuid.NewV6()
// 	if err != nil {
// 		return err
// 	}
//
// 	sql := `
// 		INSERT INTO public.gauges (id, name, value)
// 		VALUES ($1, $2, $3);`
// 	_, err = s.db.Exec(context.Background(), sql, id, name, value)
//
// 	return err
// }
//
// func (s *OrderRepo) IncCounters(counters []entity.Counter) error {
// 	if len(counters) == 0 {
// 		return nil
// 	}
//
// 	sql := strings.Builder{}
// 	sql.WriteString(`
// 		INSERT INTO public.counters (id, name, "value")
// 		VALUES`)
// 	var params []any
// 	i := 0
// 	for _, counter := range counters {
// 		if counter.ID.IsNil() {
// 			id, err := uuid.NewV6()
// 			if err != nil {
// 				return err
// 			}
// 			counter.ID = id
// 		}
// 		params = append(params, counter.ID, counter.Name, counter.Value)
// 		if i > 0 {
// 			sql.WriteString(",")
// 		}
// 		sql.WriteString(fmt.Sprintf(" ($%d, $%d, $%d)", i+1, i+2, i+3))
// 		i += 3
// 	}
// 	sql.WriteString(`
// 		ON CONFLICT (name) DO UPDATE
// 		  SET "value" = counters."value" + EXCLUDED.value;`)
//
// 	_, err := s.db.Exec(context.Background(), sql.String(), params...)
//
// 	return err
// }
//
// func (s *OrderRepo) AppendGauges(gauges []entity.Gauge) error {
// 	if len(gauges) == 0 {
// 		return nil
// 	}
//
// 	sql := strings.Builder{}
// 	sql.WriteString(`
// 		INSERT INTO public.gauges (id, name, "value")
// 		VALUES`)
// 	var params []any
// 	i := 0
// 	for _, gauge := range gauges {
// 		if gauge.ID.IsNil() {
// 			id, err := uuid.NewV6()
// 			if err != nil {
// 				return err
// 			}
// 			gauge.ID = id
// 		}
// 		params = append(params, gauge.ID, gauge.Name, gauge.Value)
// 		if i > 0 {
// 			sql.WriteString(",")
// 		}
// 		sql.WriteString(fmt.Sprintf(" ($%d, $%d, $%d)", i+1, i+2, i+3))
// 		i += 3
// 	}
//
// 	_, err := s.db.Exec(context.Background(), sql.String(), params...)
//
// 	return err
// }
//
// func (s *OrderRepo) GetCounter(name string) (int64, bool) {
// 	value := int64(0)
// 	sql := `SELECT value FROM public.counters
// 		WHERE name = $1 LIMIT 1;`
// 	err := pgxscan.Get(context.Background(), s.db, &value, sql, name)
// 	if err != nil {
// 		if !errors.Is(err, pgx.ErrNoRows) {
// 			log.WithError(err).Error("Error getting counter")
// 		}
// 		return 0, false
// 	}
//
// 	return value, true
// }
//
// func (s *OrderRepo) GetGauge(name string) ([]float64, bool) {
// 	var values []float64
// 	sql := `SELECT value FROM public.gauges
// 		WHERE name = $1
// 		ORDER BY id;`
// 	err := pgxscan.Get(context.Background(), s.db, values, sql, name)
// 	if err != nil {
// 		log.WithError(err).Error("Error getting gauges")
// 		return nil, false
// 	}
//
// 	return values, true
// }
//
// func (s *OrderRepo) GetLastGauge(name string) (float64, bool) {
// 	var value float64
// 	sql := `SELECT value FROM public.gauges
// 		WHERE name = $1
// 		ORDER BY id desc LIMIT 1;`
// 	err := pgxscan.Get(context.Background(), s.db, &value, sql, name)
// 	if err != nil {
// 		log.WithError(err).Error("Error getting last gauge")
// 		return value, false
// 	}
//
// 	return value, true
// }
//
// func (s *OrderRepo) GetCounters() map[string]int64 {
// 	var values []entity.Counter
// 	sql := `SELECT * FROM public.counters;`
// 	err := pgxscan.Select(context.Background(), s.db, &values, sql)
// 	if err != nil {
// 		log.WithError(err).Error("Error getting counters map")
// 		return nil
// 	}
//
// 	counters := make(map[string]int64, len(values))
// 	for _, counter := range values {
// 		counters[counter.Name] = counter.Value
// 	}
//
// 	return counters
// }
//
// func (s *OrderRepo) GetGauges() map[string][]float64 {
// 	var values []entity.Gauge
// 	sql := `SELECT * FROM public.gauges
// 		ORDER BY id;`
// 	err := pgxscan.Select(context.Background(), s.db, &values, sql)
// 	if err != nil {
// 		log.WithError(err).Error("Error getting gauges map")
// 		return nil
// 	}
//
// 	gauges := make(map[string][]float64)
// 	for _, gauge := range values {
// 		gauges[gauge.Name] = append(gauges[gauge.Name], gauge.Value)
// 	}
//
// 	return gauges
// }
