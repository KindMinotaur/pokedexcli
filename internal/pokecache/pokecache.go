package pokecache

import (
	"sync"
	"time"
)


type Cache struct {
	cacheEntries map[string]cacheEntry
	mu *sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val []byte
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		cacheEntries: make(map[string]cacheEntry),
		mu: &sync.Mutex{},
	}
	return c
}

func (c *Cache) Add(key string, val []byte) {
	defer c.mu.Unlock()
	c.mu.Lock()
}

func (c *Cache) Get(key string) ([]byte, bool) {
	defer c.mu.Unlock()
	c.mu.Lock()
	return nil, false
}

func (c *Cache) reapLoop(interval time.Duration) {
}