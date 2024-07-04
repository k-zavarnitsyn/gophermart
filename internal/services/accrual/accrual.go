package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/k-zavarnitsyn/gophermart/internal/config"
	"github.com/k-zavarnitsyn/gophermart/internal/utils"
	"github.com/k-zavarnitsyn/gophermart/internal/workers"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/entity"
	"github.com/k-zavarnitsyn/gophermart/pkg/domain/repository"
	log "github.com/sirupsen/logrus"
)

/*
Сервис обрабатывает запросы в том же процессе, стартуя и завершая горутины, но при превышении
допустимого числа потоков происходит переход в состояние перегрузки и заказы не обрабатываются,
пока их не подцепит функция обработки по таймеру.
TODO: Отправлять события в брокер сообщений для параллельной обработки на нескольких нодах
*/

type Service interface {
	Send(ctx context.Context, order *entity.Order) error
}

type service struct {
	cfg               *config.Accrual
	mainWorker        *workers.OverloadableWorker[*entity.Order]
	overloadCounter   int
	overloadStartTime time.Time
	processingOrders  sync.Map
	tickMu            sync.Mutex
	ticker            *time.Ticker

	orderRepo repository.Order
}

type accrualResponse struct {
	OrderNumber string  `json:"order"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual"`
}

func NewService(cfg *config.Accrual, orderRepo repository.Order) Service {
	s := &service{
		cfg:               cfg,
		orderRepo:         orderRepo,
		overloadStartTime: time.Now(),
		ticker:            time.NewTicker(cfg.PollingInterval),
	}
	s.mainWorker = workers.NewOverloadableWorker(cfg.MaxActiveWorkers, s.ProcessOrder, s.ProcessOrderOnOverload)
	s.runTicker()

	return s
}

func (s *service) runTicker() {
	go func() {
		for {
			select {
			case t := <-s.ticker.C:
				s.processTick(t)
			}
		}
	}()
}

func (s *service) Send(ctx context.Context, order *entity.Order) error {
	// для масштабируемости событию хорошо бы уходить в кафку, но пока обработка в том же процессе
	s.processingOrders.Store(order.Number, struct{}{})
	s.mainWorker.Add(ctx, order)

	return nil
}

func (s *service) ProcessOrder(ctx context.Context, order *entity.Order) {
	status, resp := s.getResponse(ctx, order)
	if status == http.StatusOK {
		if resp == nil {
			s.processError(ctx, fmt.Errorf("status is ok, but response is empty"), order, "unexpected behaviour")
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
			s.processError(ctx, fmt.Errorf("invalid status %s", resp.Status), order, "unexpected behaviour")
			return
		}
		if resp.Accrual <= 0 {
			err := s.orderRepo.SetOrderStatus(ctx, order.Number, orderStatus)
			if err != nil {
				log.WithError(err).WithField("order", order.Number).WithField("resp", resp).Error("Failed to set order status")
			}
			return
		}
		order.Status = orderStatus
		if orderStatus == entity.OrderStatusProcessed || orderStatus == entity.OrderStatusInvalid {
			order.Accrual = utils.ToPointer(resp.Accrual)
		}

		err := s.orderRepo.UpdateAttributes(ctx, order)
		if err != nil {
			log.WithError(err).WithField("order", order).WithField("resp", resp).Error("Failed to update order")
		}
	} else if status == http.StatusNoContent {
		// `204` - заказ не зарегистрирован в системе расчета.
		err := s.orderRepo.SetOrderStatus(ctx, order.Number, entity.OrderStatusInvalid)
		if err != nil {
			log.WithError(err).WithField("order", order.Number).Error("Failed to set order status")
		}
	} else {
		s.processError(ctx, fmt.Errorf("bad status %d", status), order, "")
	}
	s.processingOrders.Delete(order.Number)
}

func (s *service) ProcessOrderOnOverload(ctx context.Context, order *entity.Order) {
	s.processingOrders.Delete(order.Number)
	log.WithField("order", order.Number).Info("Processing order on overload")
	// Ничего не делаем, эти заказы обработаются позже

	s.overloadCounter++
	if s.overloadCounter >= s.cfg.OverloadReportCount {
		if float64(s.overloadCounter)/time.Since(s.overloadStartTime).Seconds() > s.cfg.OverloadReportRPS {
			log.WithField("order", order.Number).Error("Too many overload events")
		}
		s.overloadCounter = 0
		s.overloadStartTime = time.Now()
	}
}

func (s *service) getResponse(ctx context.Context, order *entity.Order) (int, *accrualResponse) {
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
	var response accrualResponse
	if err := json.Unmarshal(data, &response); err != nil {
		s.processError(ctx, err, order, "Failed to parse response body")
	}

	return resp.StatusCode, &response
}

func (s *service) processError(ctx context.Context, err error, order *entity.Order, msg string) {
	log.WithError(err).WithField("order", order.Number).Error(msg)
	// err = s.orderRepo.SetOrderStatus(ctx, order.Number, entity.OrderStatusProcessing)
	// if err != nil {
	// 	log.WithError(err).WithField("order", order.Number).Errorf("Failed to set order status")
	// }
}

func (s *service) processTick(t time.Time) {
	// only one active processor
	if !s.tickMu.TryLock() {
		return
	}
	defer s.tickMu.Unlock()

	var activeOrderNumbers []string
	s.processingOrders.Range(func(k, v interface{}) bool {
		activeOrderNumbers = append(activeOrderNumbers, k.(string))
		return true
	})
	ctx := context.Background()
	statuses := []string{string(entity.OrderStatusNew), string(entity.OrderStatusProcessing)}
	orders, err := s.orderRepo.GetOrdersByStatuses(ctx, statuses, activeOrderNumbers, s.cfg.PollingCount)
	if err != nil {
		log.WithError(err).Error("Failed to orders in accrual processing tick")
	}
	for _, order := range orders {
		err := s.Send(ctx, &order)
		if err != nil {
			log.WithError(err).WithField("order", order).Error("Failed to send order")
		}
	}
}
