package workers

import (
	"context"
	"sync"
	"sync/atomic"
)

type Handler[T any] func(ctx context.Context, arg T)

type OverloadableWorker[T any] struct {
	maxWorkers      int32
	workersNum      atomic.Int32
	handler         Handler[T]
	overloadHandler Handler[T]
	wg              sync.WaitGroup
}

func NewOverloadableWorker[T any](maxWorkers int, handler Handler[T], onOverload Handler[T]) *OverloadableWorker[T] {
	return &OverloadableWorker[T]{
		maxWorkers:      int32(maxWorkers),
		handler:         handler,
		overloadHandler: onOverload,
	}
}

func (p *OverloadableWorker[T]) Add(ctx context.Context, arg T) {
	if p.workersNum.Load() < p.maxWorkers {
		p.wg.Add(1)
		p.workersNum.Add(1)
		go func() {
			defer func() {
				p.workersNum.Add(-1)
				p.wg.Done()
			}()
			p.handler(ctx, arg)
		}()
	} else {
		p.overloadHandler(ctx, arg)
	}
}

func (p *OverloadableWorker[T]) Wait() {
	p.wg.Wait()
}
