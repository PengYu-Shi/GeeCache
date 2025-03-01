package gee

import (
	"GeeCacheNode/gee/byteview"
	"time"

	"GeeCacheNode/gee/lru"
	"GeeCacheNode/gee/singleflight"
	"GeeCacheNode/gee/snapshot"
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, duration time.Duration, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()

	loadCache := lru.NewCache(cacheBytes, nil)
	newMgr := snapshot.NewManager(loadCache)
	now := time.Now().Unix()
	fmt.Println("-----------------------------------Try To Load Old Data-----------------------------------")
	sum, err := newMgr.Load()
	newMgr.AutoSnapshot(duration)
	if err != nil {
		log.Println("[Snap Manager] Load Old Data error: ", err)
		log.Println("[Snap Manager] Old data not found! Please check path! (Ignore when there is no old data)")

	}
	log.Println("[Snap Manager] Load Old Done! Load old data sum: ", sum, ", Time consumption:", time.Now().Unix()-now)

	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
			lru:        loadCache,
		},
		loader: &singleflight.Group{},
	}

	groups[name] = g
	return g
}

func GetGroup(name string) (*Group, error) {
	mu.RLock()
	defer mu.RUnlock()
	if g, ok := groups[name]; !ok {
		return g, fmt.Errorf("No such group: %v", name)
	} else {
		return g, nil
	}
}

func (g *Group) Get(key string) (byteview.ByteView, error) {
	if key == "" {
		return byteview.ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	return g.load(key)
}

func (g *Group) getLocally(key string) (byteview.ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return byteview.ByteView{}, err
	}
	value := byteview.ByteView{B: byteview.CloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value byteview.ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) load(key string) (value byteview.ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(byteview.ByteView), nil
	}
	return
}
