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
		total := 100
		pool := workers.Start(3)
		for i := 0; i < total; i++ {
			pool.Run(func() {
				if _, exists := results.Load(i); exists {
					t.Error("duplicated results")
				}
				results.Store(i, struct{}{})
				counter.Add(1)
			})
		}
		pool.Close()

		assert.Equal(t, int64(total), counter.Load())
	})
}
