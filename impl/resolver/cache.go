package resolver

import (
	"sync"
	"time"
)

type cacheEntry struct {
	cfg     ResolvedConfig
	meta    Meta
	expires time.Time
}

type cacheStore struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
	now     func() time.Time
}

func newCacheStore() *cacheStore {
	return &cacheStore{
		entries: make(map[string]cacheEntry),
		now:     time.Now,
	}
}

func (c *cacheStore) get(key string) (ResolvedConfig, Meta, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return ResolvedConfig{}, Meta{}, false
	}

	if c.now().After(entry.expires) {
		delete(c.entries, key)
		return ResolvedConfig{}, Meta{}, false
	}

	return entry.cfg, entry.meta, true
}

func (c *cacheStore) set(key string, cfg ResolvedConfig, meta Meta, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		cfg:     cfg,
		meta:    meta,
		expires: c.now().Add(ttl),
	}
}

func (c *cacheStore) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
}

var defaultCache = newCacheStore()

// ClearCache clears the in-memory config cache.
func ClearCache() {
	defaultCache.clear()
}
