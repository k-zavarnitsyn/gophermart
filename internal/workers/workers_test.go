package workers_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/k-zavarnitsyn/gophermart/internal/workers"
	"github.com/stretchr/testify/assert"
)

func Test_workers_Pool(t *testing.T) {
	t.Run("Add all tasks", func(t *testing.T) {
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

func Test_workers_OverloadableWorker(t *testing.T) {
	t.Run("Add all tasks", func(t *testing.T) {
		results := &sync.Map{}
		counter := atomic.Int64{}
		overloadCounter := atomic.Int64{}
		total := 10000
		pool := workers.NewOverloadableWorker[int](100, func(ctx context.Context, arg int) {
			if _, exists := results.Load(arg); exists {
				t.Error("duplicated results")
			}
			results.Store(arg, struct{}{})
			counter.Add(1)
		}, func(ctx context.Context, arg int) {
			if _, exists := results.Load(arg); exists {
				t.Error("duplicated results")
			}
			results.Store(arg, struct{}{})
			overloadCounter.Add(1)
		})
		ctx := context.Background()
		for i := 0; i < total; i++ {
			pool.Add(ctx, i)
		}
		pool.Wait()

		t.Logf("normal: %d", counter.Load())
		t.Logf("overload: %d", overloadCounter.Load())
		assert.Equal(t, int64(total), counter.Load()+overloadCounter.Load())
	})
}
