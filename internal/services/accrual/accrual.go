package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
	"github.com/k-zavarnitsyn/gophermart/internal/workers"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/repository"
	log "github.com/sirupsen/logrus"
)

/*
Сервис обрабатывает запросы в том же процессе, стартуя и завершая горутины, но при превышении
допустимого числа потоков происходит переход в состояние перегрузки и события отправляются в пул
с фиксированным небольшим числом воркеров.
TODO: Отправлять события в брокер сообщений для параллельной обработки на нескольких нодах
*/

type Service interface {
	Send(ctx context.Context, event *AccrualEvent) error
}

type service struct {
	cfg  *config.Accrual
	pool *workers.Pool

	orderRepo repository.Order
}

func NewService(cfg *config.Accrual, orderRepo repository.Order) Service {
	return &service{
		cfg:       cfg,
		pool:      workers.Start(cfg.PoolSize),
		orderRepo: orderRepo,
	}
}

func (s *service) Send(ctx context.Context, event *AccrualEvent) error {
	// для масштабируемости событию хорошо бы уходить в кафку, но пока обработка в том же процессе
	s.pool.Run(func() {
		status, resp := s.getResponse(ctx, event.Order)
		if status == http.StatusOK {
			if resp == nil {
				s.processError(ctx, fmt.Errorf("status is ok, but response is empty"), event.Order, "unexpected behaviour")
				return
			}
			var orderStatus entity.OrderStatus
			switch resp.Status {
			case `REGISTERED`: // — заказ зарегистрирован, но не начисление не рассчитано;
				orderStatus = entity.OrderStatusNew
			case `PROCESSING`: // — расчёт начисления в процессе;
				orderStatus = entity.OrderStatusProcessing
			case `INVALID`: // — заказ не принят к расчёту, и вознаграждение не будет начислено;
				orderStatus = entity.OrderStatusInvalid
			case `PROCESSED`: // — расчёт начисления окончен;
				orderStatus = entity.OrderStatusProcessed
			default:
				s.processError(ctx, fmt.Errorf("invalid status %s", resp.Status), event.Order, "unexpected behaviour")
				return
			}
			if resp.Accrual <= 0 {
				err := s.orderRepo.SetOrderStatus(ctx, event.Order, orderStatus)
				if err != nil {
					log.WithError(err).WithField("order", event.Order.Number).WithField("resp", resp).Error("Failed to set order status")
				}
				return
			}
			event.Order.Status = orderStatus
			event.Order.Accrual = utils.ToPointer(resp.Accrual)

			err := s.orderRepo.UpdateAttributes(ctx, event.Order)
			if err != nil {
				log.WithError(err).WithField("order", event.Order).WithField("resp", resp).Error("Failed to update order")
			}
		} else if status == http.StatusNoContent {
			// `204` - заказ не зарегистрирован в системе расчета.
			err := s.orderRepo.SetOrderStatus(ctx, event.Order, entity.OrderStatusInvalid)
			if err != nil {
				log.WithError(err).WithField("order", event.Order.Number).Error("Failed to set order status")
			}
		} else {
			s.processError(ctx, fmt.Errorf("bad status %d", status), event.Order, "")
		}
	})

	return nil
}

func (s *service) getResponse(ctx context.Context, order *entity.Order) (int, *AccrualResponse) {
	url := fmt.Sprintf("%s/api/orders/%s", s.cfg.AccrualSystemAddress, order.Number)
	resp, err := http.Get(url) //nolint:gosec // Variable url is fine here (number is validated)
	if err != nil {
		s.processError(ctx, err, order, "Failed to send request to accrual")
		return 0, nil
	}
	defer utils.CloseWithLogging(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		s.processError(ctx, err, order, "Failed to read response body")
		return 0, nil
	}
	var response AccrualResponse
	if err := json.Unmarshal(data, &response); err != nil {
		s.processError(ctx, err, order, "Failed to parse response body")
	}

	return resp.StatusCode, &response
}

func (s *service) processError(ctx context.Context, err error, order *entity.Order, msg string) {
	log.WithError(err).WithField("order", order.Number).Error(msg)
	err = s.orderRepo.SetOrderStatus(ctx, order, entity.OrderStatusProcessing)
	if err != nil {
		log.WithError(err).WithField("order", order.Number).Errorf("Failed to set order status")
	}
}
