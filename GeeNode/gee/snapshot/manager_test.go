package snapshot

import (
	"GeeCacheNode/gee/byteview"
	"GeeCacheNode/gee/lru"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testSnapshotDir = "./testdata/snapshots"

func setupTestEnv(t *testing.T) (*lru.Cache, func()) {
	// 创建测试缓存实例
	c := lru.NewCache(1<<30, nil) // 1GB容量

	// 填充测试数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		value := bytes.Repeat([]byte{byte(i)}, 1024) // 1KB数据
		c.Add(key, byteview.ByteView{B: value})      // TTL 1-10分钟
	}

	// 创建临时目录
	os.MkdirAll(testSnapshotDir, 0755)

	// 返回清理函数
	return c, func() {
		os.RemoveAll(testSnapshotDir)
	}
}

func TestSnapshotLifecycle(t *testing.T) {
	c, cleanup := setupTestEnv(t)
	defer cleanup()

	mgr := NewManager(c)

	// 测试快照生成
	t.Run("GenerateSnapshot", func(t *testing.T) {
		err := mgr.CreateSnapshot()
		assert.NoError(t, err)

		// 验证元数据文件
		metaPath := filepath.Join(testSnapshotDir, "current/meta.json")
		metaData, err := ioutil.ReadFile(metaPath)
		assert.NoError(t, err)

		var meta Meta
		assert.NoError(t, json.Unmarshal(metaData, &meta))
		assert.Equal(t, 1, meta.Version)
		assert.Equal(t, 1000, meta.ItemCount)
		assert.Equal(t, "zstd", meta.Compression)
	})

	// 测试快照加载
	t.Run("LoadSnapshot", func(t *testing.T) {
		// 创建新缓存实例
		newCache := lru.NewCache(1<<30, nil)
		newMgr := NewManager(newCache)

		// 加载快照
		err := newMgr.Load()
		assert.NoError(t, err)

		// 验证数据完整性
		assert.Equal(t, 1000, newCache.Len())

		// 随机抽查数据
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key%d", i)
			_, ok := newCache.Get(key)
			assert.True(t, ok)
			//assert.Len(t, val.ByteSlice(), 1024)
		}
	})
}

func TestCorruptedSnapshot(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// 创建损坏的快照
	corruptDir := filepath.Join(testSnapshotDir, "corrupt")
	os.MkdirAll(corruptDir, 0755)

	// 写入无效数据
	ioutil.WriteFile(filepath.Join(corruptDir, "data.zst"), []byte("invalid"), 0644)
	ioutil.WriteFile(filepath.Join(corruptDir, "meta.json"), []byte(`{"version":1}`), 0644)

	t.Run("LoadCorruptedData", func(t *testing.T) {
		newCache := lru.NewCache(1<<30, nil)
		mgr := NewManager(newCache)

		// 应返回错误
		err := mgr.Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据文件校验失败")
	})
}

func TestVersionCompatibility(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// 创建不兼容版本
	invalidMeta := Meta{
		Version: 99,
		Created: time.Now(),
	}
	metaData, _ := json.Marshal(invalidMeta)
	os.MkdirAll(filepath.Join(testSnapshotDir, "current"), 0755)
	ioutil.WriteFile(filepath.Join(testSnapshotDir, "current/meta.json"), metaData, 0644)

	t.Run("InvalidVersion", func(t *testing.T) {
		newCache := lru.NewCache(1<<30, nil)
		mgr := NewManager(newCache)

		err := mgr.Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "不支持的快照版本")
	})
}

func TestParallelLoading(t *testing.T) {
	c, cleanup := setupTestEnv(t)
	defer cleanup()

	// 生成包含10万条数据的大快照
	for i := 0; i < 100000; i++ {
		key := fmt.Sprintf("key%d", i)
		c.Add(key, byteview.ByteView{B: []byte("value")})
	}

	mgr := NewManager(c)
	assert.NoError(t, mgr.CreateSnapshot())

	t.Run("ConcurrentLoad", func(t *testing.T) {
		newCache := lru.NewCache(1<<30, nil)
		newMgr := NewManager(newCache)

		done := make(chan struct{})
		go func() {
			defer close(done)
			assert.NoError(t, newMgr.Load())
		}()

		// 并行读取测试
		for i := 0; i < 1000; i++ {
			go func(i int) {
				key := fmt.Sprintf("key%d", i)
				_, ok := newCache.Get(key)
				assert.True(t, ok)
			}(i)
		}
		<-done
	})
}
