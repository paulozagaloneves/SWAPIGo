package cache

import (
	"sync"
	"time"
)

type entry struct {
	status int
	ctype  string
	body   []byte
	exp    time.Time
}

type Cache interface {
	Get(key string) (status int, ctype string, body []byte, ok bool)
	Set(key string, status int, ctype string, body []byte)
}

type ttlCache struct {
	mu    sync.RWMutex
	items map[string]entry
	cap   int
	ttl   time.Duration
	order []string // naive FIFO order for eviction
}

func New(capacity int, ttl time.Duration) Cache {
	return &ttlCache{
		items: make(map[string]entry, capacity),
		cap:   capacity,
		ttl:   ttl,
		order: make([]string, 0, capacity),
	}
}

func (c *ttlCache) Get(key string) (int, string, []byte, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return 0, "", nil, false
	}
	if time.Now().After(e.exp) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return 0, "", nil, false
	}
	return e.status, e.ctype, append([]byte(nil), e.body...), true
}

func (c *ttlCache) Set(key string, status int, ctype string, body []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.items) >= c.cap {
		// evict oldest (FIFO)
		if len(c.order) > 0 {
			oldest := c.order[0]
			c.order = c.order[1:]
			delete(c.items, oldest)
		}
	}
	c.items[key] = entry{status: status, ctype: ctype, body: append([]byte(nil), body...), exp: time.Now().Add(c.ttl)}
	c.order = append(c.order, key)
}
