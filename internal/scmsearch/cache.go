package scmsearch

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"sync/atomic"

	"github.com/mesdx/cli/internal/treesitter"
)

// QueryCache is a thread-safe LRU cache for compiled Tree-sitter queries.
type QueryCache struct {
	mu       sync.Mutex
	entries  map[string]*queryCacheEntry
	order    []string // oldest first
	capacity int
	hits     atomic.Int64
	misses   atomic.Int64
}

type queryCacheEntry struct {
	query *treesitter.Query
	lang  *treesitter.Language
}

// NewQueryCache creates a cache with the given capacity.
func NewQueryCache(capacity int) *QueryCache {
	if capacity < 1 {
		capacity = 64
	}
	return &QueryCache{
		entries:  make(map[string]*queryCacheEntry, capacity),
		order:    make([]string, 0, capacity),
		capacity: capacity,
	}
}

func queryKey(langName, querySrc string) string {
	h := sha256.Sum256([]byte(langName + "\x00" + querySrc))
	return hex.EncodeToString(h[:])
}

// Get returns a cached compiled query or nil.
func (c *QueryCache) Get(langName, querySrc string) *treesitter.Query {
	key := queryKey(langName, querySrc)
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[key]; ok {
		c.moveToBack(key)
		c.hits.Add(1)
		return e.query
	}
	c.misses.Add(1)
	return nil
}

// Put stores a compiled query in the cache.
func (c *QueryCache) Put(langName, querySrc string, lang *treesitter.Language, q *treesitter.Query) {
	key := queryKey(langName, querySrc)
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.entries[key]; ok {
		c.moveToBack(key)
		return
	}
	if len(c.entries) >= c.capacity {
		c.evictOldest()
	}
	c.entries[key] = &queryCacheEntry{query: q, lang: lang}
	c.order = append(c.order, key)
}

// Stats returns (hits, misses).
func (c *QueryCache) Stats() (int64, int64) {
	return c.hits.Load(), c.misses.Load()
}

func (c *QueryCache) moveToBack(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, key)
			return
		}
	}
}

func (c *QueryCache) evictOldest() {
	if len(c.order) == 0 {
		return
	}
	oldest := c.order[0]
	c.order = c.order[1:]
	if e, ok := c.entries[oldest]; ok {
		e.query.Close()
		delete(c.entries, oldest)
	}
}

// Close frees all cached queries.
func (c *QueryCache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, e := range c.entries {
		e.query.Close()
	}
	c.entries = make(map[string]*queryCacheEntry)
	c.order = c.order[:0]
}
