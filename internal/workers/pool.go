package workers

import (
	"sync"
)

type Task func()

type Waiter interface {
	Wait()
}

type Pool struct {
	size    int
	jobs    chan Task
	wg      sync.WaitGroup
	started bool
	mu      sync.Mutex
}

func NewPool(size int) *Pool {
	return &Pool{
		size: size,
	}
}

func Start(poolSize int) *Pool {
	p := NewPool(poolSize)
	p.Start()

	return p
}

func (p *Pool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.start()
}

func (p *Pool) start() {
	if p.started {
		return
	}
	p.jobs = make(chan Task)
	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for job := range p.jobs {
				job()
			}
		}()
	}
	p.started = true
}

func (p *Pool) Run(task Task) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		p.start()
	}
	p.jobs <- task
}

func (p *Pool) Wait() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started {
		return
	}
	close(p.jobs)
	// ждем завершения воркеров, иначе send on closed channel
	p.wg.Wait()
	p.started = false
}
