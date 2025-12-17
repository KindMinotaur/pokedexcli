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
	go c.reapLoop(interval)
	return c
}

func (c *Cache) Add(key string, val []byte) {
	defer c.mu.Unlock()
	c.mu.Lock()
	c.cacheEntries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	defer c.mu.Unlock()
	c.mu.Lock()
	if entry, ok := c.cacheEntries[key]; ok {
		return entry.val, true
	}else {
		return nil, false
	}
}

func (c *Cache) reapLoop(interval time.Duration) {
	defer c.mu.Unlock()
	c.mu.Lock()
	ticker := time.NewTicker(interval)
	for range ticker.C {
		<-ticker.C
		now := time.Now()
		for key, entry := range c.cacheEntries {
			if now.Sub(entry.createdAt) > interval {
				delete(c.cacheEntries, key)
			}
		}
	}
}