package workers_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/k-zavarnitsyn/gophermart/internal/workers"
	"github.com/stretchr/testify/assert"
)

func Test_workers_Pool(t *testing.T) {
	t.Run("Run all tasks", func(t *testing.T) {
		results := &sync.Map{}
		counter := atomic.Int64{}
		total := 1000
		pool := workers.NewPool(10)
		for i := 0; i < total; i++ {
			pool.Run(func() {
				if _, exists := results.Load(i); exists {
					t.Error("duplicated results")
				}
				results.Store(i, struct{}{})
				counter.Add(1)
			})
		}
		pool.Wait()

		assert.Equal(t, int64(total), counter.Load())
	})
}

func Test_workers_ActivePool(t *testing.T) {
	t.Run("Run all tasks", func(t *testing.T) {
		results := &sync.Map{}
		counter := atomic.Int64{}
		total := 10000
		pool := workers.NewActivePool(&workers.ActivePoolConfig{
			MaxActiveWorkers: 100,
			OverloadPoolSize: 10,
		})
		for i := 0; i < total; i++ {
			pool.Run(func() {
				if _, exists := results.Load(i); exists {
					t.Error("duplicated results")
				}
				results.Store(i, struct{}{})
				counter.Add(1)
			})
		}
		pool.Wait()

		assert.Equal(t, int64(total), counter.Load())
	})
}
