package gee

import "C"
import (
	"GeeCacheNode/gee/byteview"
	"GeeCacheNode/gee/lru"
	"GeeCacheNode/gee/snapshot"
	"sync"
	"time"
)

type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value byteview.ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.NewCache(c.cacheBytes, nil)
		snapManager := snapshot.NewManager(c.lru)
		//snapManager.AutoSnapshot(time.Minute * 2)
		snapManager.AutoSnapshot(time.Second * 10)

	}

	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value byteview.ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(byteview.ByteView), ok
	}

	return

}
