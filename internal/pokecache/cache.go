package pokecache

import (
	"sync"
	"time"
)

// cacheEntry represents a single cached item.
type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

// Cache is a simple thread-safe in-memory cache with time-based reaping.
type Cache struct {
	mu       sync.Mutex
	entries  map[string]cacheEntry
	interval time.Duration
}

// NewCache creates a cache and starts the reap loop which deletes entries older than interval.
func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entries:  make(map[string]cacheEntry),
		interval: interval,
	}
	go c.reapLoop()
	return c
}

// Add stores a value in the cache under the given key.
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// store a copy of val to avoid callers mutating our internal slice
	v := make([]byte, len(val))
	copy(v, val)
	c.entries[key] = cacheEntry{
		createdAt: time.Now().UTC(),
		val:       v,
	}
}

// Get returns a copy of the cached value for key and true if present.
// Returns nil,false if not found.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	// return a copy so callers cannot mutate internal slice
	v := make([]byte, len(e.val))
	copy(v, e.val)
	return v, true
}

// reapLoop periodically removes entries older than the cache interval.
func (c *Cache) reapLoop() {
	if c.interval <= 0 {
		return
	}
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		cutoff := time.Now().UTC().Add(-c.interval)
		for k, e := range c.entries {
			if e.createdAt.Before(cutoff) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}
