package lru

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	mu        sync.RWMutex
	maxsize   int64
	tmpsize   int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type Content struct {
	Key         string
	Value       Value
	CreateAt    int64
	ExpiredTime int64 //default 30 second
}

type Value interface {
	Len() int
}

func NewCache(maxsize int64, onEvicted func(string, Value)) *Cache {
	cacheT := &Cache{
		maxsize:   maxsize,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
	return cacheT
}

func (c *Cache) RemoveTheOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*Content)
		delete(c.cache, kv.Key)
		c.tmpsize -= int64(len(kv.Key)) + int64(kv.Value.Len())
		fmt.Println("lru remove！!")
		if c.OnEvicted != nil {
			c.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (c *Cache) RemoveExpired() {
	for last := c.ll.Back(); last != nil; last = last.Prev() {
		kv := last.Value.(*Content)
		if kv.CreateAt+kv.ExpiredTime > time.Now().Unix() {
			break
		}
		c.ll.Remove(last)
		delete(c.cache, kv.Key)
		fmt.Println("lru expired！!")
		c.tmpsize -= int64(len(kv.Key)) + int64(kv.Value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*Content)
		return kv.Value, true
	} else {
		return nil, false
	}
}

func (c *Cache) Add(key string, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*Content)
		kv.CreateAt = time.Now().Unix()
		kv.Value = value
	} else {
		ele := c.ll.PushFront(&Content{
			Key:         key,
			Value:       value,
			CreateAt:    time.Now().Unix(),
			ExpiredTime: time.Now().Unix() + 3600,
		})
		fmt.Println("key: ", key, "created at : ", ele.Value.(*Content).CreateAt)
		c.cache[key] = ele
		c.tmpsize += int64(len(key)) + int64(value.Len())
	}
	c.RemoveExpired()
	if c.tmpsize > c.maxsize {
		c.RemoveTheOldest()
	}

}

func (c *Cache) Len() int {
	return c.ll.Len()
}

// 获取只读视图
func (c *Cache) GetReadOnlyView() []*Content {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items := make([]*Content, 0, c.ll.Len())
	for e := c.ll.Front(); e != nil; e = e.Next() {
		items = append(items, &Content{
			Key:         e.Value.(*Content).Key,
			Value:       e.Value.(*Content).Value,
			CreateAt:    e.Value.(*Content).CreateAt,
			ExpiredTime: e.Value.(*Content).ExpiredTime,
		})
	}
	return items
}
