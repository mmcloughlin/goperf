package sem

import (
	"context"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

type Set struct {
	s  map[int]bool
	mu sync.Mutex
}

func NewSet() *Set {
	return &Set{
		s: map[int]bool{},
	}
}

// Add x to the set. Return whether it was already there.
func (s *Set) Add(x int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	in := s.s[x]
	s.s[x] = true
	return in
}

// Remove from the set.
func (s *Set) Remove(x int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.s, x)
}

func Hammer(t *testing.T, p *Pool, s *Set, n, loops int) {
	const maxsleep = 200 * time.Microsecond

	for i := 0; i < loops; i++ {
		// Acquire up to n items.
		a := 1 + rand.Intn(n)
		items, err := p.Acquire(context.Background(), a)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != a {
			t.Fatalf("got %d items; expected %d", len(items), a)
		}

		// Add them all to the set.
		for _, item := range items {
			used := s.Add(item.(int))
			if used {
				t.Fatalf("non-exclusive access to %d", item.(int))
			}
		}

		// Hold items for a while.
		time.Sleep(time.Duration(rand.Int63n(int64(maxsleep/time.Nanosecond))) * time.Nanosecond)

		// Remove from set.
		for _, item := range items {
			s.Remove(item.(int))
		}

		// Return to pool.
		p.Release(items)
	}
}

func TestPool(t *testing.T) {
	t.Parallel()

	// Pool of items.
	n := 24
	items := make([]interface{}, n)
	for i := 0; i < n; i++ {
		items[i] = i
	}
	pool := NewPool(items...)

	// Setup GOMAXPROCS goroutines accessing the pool.
	m := runtime.GOMAXPROCS(0)
	loops := 10000 / m

	s := NewSet()
	var wg sync.WaitGroup
	wg.Add(m)
	for i := 0; i < m; i++ {
		go func() {
			defer wg.Done()
			Hammer(t, pool, s, n, loops)
		}()
	}
	wg.Wait()
}
