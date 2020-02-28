// Package sem provides safe concurrent access to a bounded pool of resources.
package sem

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

// Pool provides concurrent access to a fixed bounded pool of resources
type Pool struct {
	items []interface{}
	mu    sync.Mutex // protects items

	sem *semaphore.Weighted
}

// NewPool builds a pool providing concurrent access to the given items.
func NewPool(items ...interface{}) *Pool {
	return &Pool{
		items: items,
		sem:   semaphore.NewWeighted(int64(len(items))),
	}
}

// Acquire n items from the pool. Blocks until the requested items are available
// or the context is cancelled.
func (p *Pool) Acquire(ctx context.Context, n int) ([]interface{}, error) {
	// Acquire weight n from the semaphore.
	if err := p.sem.Acquire(ctx, int64(n)); err != nil {
		return nil, err
	}

	// Take n items from the pool.
	p.mu.Lock()
	if len(p.items) < n {
		panic("assertion failure: not enough items in pool")
	}
	items := p.items[:n]
	p.items = p.items[n:]
	p.mu.Unlock()

	return items, nil
}

// Release returns items to the pool.
func (p *Pool) Release(items []interface{}) {
	// Return items to the pool.
	p.mu.Lock()
	p.items = append(p.items, items...)
	p.mu.Unlock()

	// Release weight in the semaphore.
	p.sem.Release(int64(len(items)))
}
