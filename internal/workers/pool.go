package workers

import "sync"

type Pool struct {
	size int
	jobs chan func()
	wg   sync.WaitGroup
}

func newPool(size int) *Pool {
	return &Pool{
		size: size,
		jobs: make(chan func()),
	}
}

func Start(poolSize int) *Pool {
	p := newPool(poolSize)
	p.Start()

	return p
}

func (p *Pool) Start() {
	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for job := range p.jobs {
				job()
			}
		}()
	}
}

func (p *Pool) Run(task func()) {
	p.jobs <- task
}

func (p *Pool) Close() {
	close(p.jobs)
	// ждем завершения воркеров, иначе send on closed channel
	p.wg.Wait()
}
