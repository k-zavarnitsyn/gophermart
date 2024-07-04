package workers

import (
	"sync"
	"sync/atomic"
)

type OverloadHandler func(task Task)

type ActivePoolConfig struct {
	MaxActiveWorkers int
	OverloadPoolSize int
}

// ActivePool Обрабатывает запросы в том же процессе, стартуя и завершая горутины, но при превышении
// допустимого числа потоков происходит переход в состояние перегрузки и события отправляются в пул
// с фиксированным небольшим числом воркеров.
type ActivePool struct {
	*OverloadableWorker
	pool *Pool
}

type OverloadableWorker struct {
	maxWorkers      int32
	workersNum      atomic.Int32
	overloadHandler OverloadHandler
	wg              sync.WaitGroup
}

func NewOverloadableWorker(maxWorkers int32, onOverload OverloadHandler) *OverloadableWorker {
	return &OverloadableWorker{
		maxWorkers:      maxWorkers,
		overloadHandler: onOverload,
	}
}

func NewActivePool(cfg *ActivePoolConfig) *ActivePool {
	p := NewPool(cfg.OverloadPoolSize)
	return &ActivePool{
		OverloadableWorker: NewOverloadableWorker(int32(cfg.MaxActiveWorkers), p.Run),
		pool:               p,
	}
}

func (p *OverloadableWorker) Run(task Task) {
	if p.workersNum.Load() < p.maxWorkers {
		p.wg.Add(1)
		p.workersNum.Add(1)
		go func() {
			defer func() {
				p.workersNum.Add(-1)
				p.wg.Done()
			}()
			task()
		}()
	} else {
		p.overloadHandler(task)
	}
}

func (p *OverloadableWorker) Wait() {
	p.wg.Wait()
}

func (p *ActivePool) Wait() {
	p.OverloadableWorker.Wait()
	p.pool.Wait()
}
