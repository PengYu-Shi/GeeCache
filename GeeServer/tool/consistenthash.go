package tool

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type Hash func(data []byte) uint32

type Map struct {
	hash      Hash
	replicas  int
	keys      []int
	ipHashMap map[int]string
}

var (
	HashMap = &Map{}
	mu      sync.Mutex
)

func New(replicas int, fn Hash) *Map {
	HashMap = &Map{
		replicas:  replicas,
		hash:      fn,
		ipHashMap: make(map[int]string),
	}
	if HashMap.hash == nil {
		HashMap.hash = crc32.ChecksumIEEE
	}
	return HashMap
}

func (m *Map) Add(keys []string) {
	mu.Lock()
	defer mu.Unlock()
	for _, key := range keys {
		fmt.Println("Add node: ", key, " to hash")
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.ipHashMap[hash] = key
		}
	}
}
func (m *Map) Delete(key string) {
	fmt.Println("Delete node: ", key, " to hash")
	mu.Lock()
	defer mu.Unlock()
	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
		m.keys = append(m.keys, hash)
		delete(m.ipHashMap, hash)
	}
}

func (m *Map) Get(key string) (ip string, err error) {
	if len(m.keys) == 0 {
		err = fmt.Errorf("please enter an right key")
		return "", err
	}
	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.ipHashMap[m.keys[idx%len(m.keys)]], nil
}
