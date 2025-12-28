package cache

import (
	"context"
	"sync"
	"time"
)

type item[V any] struct {
	val      V
	expires  time.Time
	hasValue bool
}

type TTLCache[K comparable, V any] struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[K]item[V]
}

func NewTTLCache[K comparable, V any](ttl time.Duration) *TTLCache[K, V] {
	return &TTLCache[K, V]{
		ttl:   ttl,
		items: make(map[K]item[V]),
	}
}

func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()

	var zero V
	if !ok || !it.hasValue {
		return zero, false
	}

	if time.Now().After(it.expires) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return zero, false
	}

	return it.val, true
}

func (c *TTLCache[K, V]) Set(key K, val V) {
	c.mu.Lock()
	c.items[key] = item[V]{
		val:      val,
		expires:  time.Now().Add(c.ttl),
		hasValue: true,
	}
	c.mu.Unlock()
}

func (c *TTLCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *TTLCache[K, V]) StartJanitor(ctx context.Context, interval time.Duration, onSweep func(removed int, size int)) {
	if interval <= 0 {
		interval = 1 * time.Second
	}

	t := time.NewTicker(interval)

	go func() {
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				now := time.Now()
				removed := 0
				c.mu.Lock()
				for k, it := range c.items {
					if it.hasValue && now.After(it.expires) {
						delete(c.items, k)
						removed++
					}
				}
				size := len(c.items)
				c.mu.Unlock()
				if onSweep != nil && removed > 0 {
					onSweep(removed, size)
				}
			}
		}
	}()
}
