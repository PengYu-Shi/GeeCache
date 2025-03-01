package snapshot

import (
	"GeeCacheNode/gee/byteview"
	"GeeCacheNode/gee/lru"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/klauspost/compress/zstd"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	snapshotBase  = "/data/127_0_0_1_9003/cache/snapshots"
	currentLink   = snapshotBase + "/current"
	tmpDirPrefix  = snapshotBase + "/tmp/"
	keepSnapshots = 3
)

// var snapshotBase string
//
// var (
//
//	currentLink   = snapshotBase + "/current"
//	tmpDirPrefix  = snapshotBase + "/tmp/"
//	keepSnapshots = 3
//
// )
type SnapshotManager struct {
	cache     *lru.Cache
	enc       *zstd.Encoder
	dec       *zstd.Decoder
	snapshots []string
	mu        sync.Mutex
}

//
//func SetRoot(path string) {
//	if path != "" {
//		snapshotBase = path
//	}
//}

func NewManager(c *lru.Cache) *SnapshotManager {
	enc, _ := zstd.NewWriter(nil)
	dec, _ := zstd.NewReader(nil)
	return &SnapshotManager{
		cache: c,
		enc:   enc,
		dec:   dec,
	}
}

func (m *SnapshotManager) AutoSnapshot(interval time.Duration) {
	log.Println("[Snap Manager] Will try to create snap")
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			log.Println("[Snap Manager] Try to create snap")
			err := m.CreateSnapshot()
			if err != nil {
				log.Println("[Snap Manager] Create snap error: ", err)
				continue
			}
			log.Println("[Snap Manager] Create snap success")
		}
	}()
}

func (m *SnapshotManager) CreateSnapshot() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tmpDir := tmpDirPrefix + time.Now().Format("20060102-150405")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		log.Println("[Snap Manager(CreateSnapshot(MkdirAll))] occur error")
		return err
	}

	view := m.cache.GetReadOnlyView()
	filtered := filterAndSort(view)

	dataFile := filepath.Join(tmpDir, "data.zst")
	if err := m.encodeToFile(filtered, dataFile); err != nil {
		log.Println("[Snap Manager(CreateSnapshot(encodeToFile))] occur error")
		return err
	}

	meta := Meta{
		Version:     1,
		Created:     time.Now(),
		ItemCount:   len(filtered),
		DataHash:    computeHash(dataFile),
		Compression: "zstd",
	}
	if err := saveMeta(meta, filepath.Join(tmpDir, "meta.json")); err != nil {
		log.Println("[Snap Manager(CreateSnapshot(saveMeta))] occur error")
		return err
	}

	return atomicSwitch(tmpDir)
}

func (m *SnapshotManager) Load() (int, error) {
	dataFile := filepath.Join(currentLink, "data.zst")
	metaFile := filepath.Join(currentLink, "meta.json")

	meta, err := validate(metaFile, dataFile)
	if err != nil {
		log.Println("[Snap Manager]: Load error when validate")
		return 0, err
	}

	data, err := m.decodeFile(dataFile)
	if err != nil {
		log.Println("[Snap Manager]: Load error when decodeFile")
		return 0, err
	}
	err = m.parallelLoad(data, meta.ItemCount)
	if err != nil {
		return 0, err
	}
	return meta.ItemCount, nil
}

func (m *SnapshotManager) decodeFile(path string) (*ColumnData, error) {
	// 1. 读取压缩数据
	compressed, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取数据文件失败: %w", err)
	}

	// 2. 使用解码器解码
	decoder := NewDecoder()
	defer decoder.Close()

	return decoder.Decode(compressed)
}

func (m *SnapshotManager) encodeToFile(items []*lru.Content, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 1. 转换为列式数据
	columns := toColumns(items)

	// 2. 序列化为字节流
	var buf bytes.Buffer
	if err := m.encodeColumns(&buf, columns); err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	originalSize := buf.Len()
	encoder := NewEncoder()
	compressed, err := encoder.Encode(columns)
	// 3. 压缩并写入
	//compressed := m.enc.EncodeAll(buf.Bytes(), nil)
	if _, err := file.Write(compressed); err != nil {
		return fmt.Errorf("写入失败: %w", err)
	}
	compressedSize := len(compressed)
	compressionRate := (float64(originalSize) - float64(compressedSize)) / float64(originalSize) * 100
	log.Printf("压缩率: %.2f%% (原始大小: %d 字节, 压缩后大小: %d 字节)", compressionRate, originalSize, compressedSize)
	return nil
}

func (m *SnapshotManager) encodeColumns(buf *bytes.Buffer, columns *ColumnData) error {
	// 写入 Keys 的长度
	if err := binary.Write(buf, binary.LittleEndian, int32(len(columns.Keys))); err != nil {
		return err
	}
	for _, key := range columns.Keys {
		// 写入每个 Key 的长度
		if err := binary.Write(buf, binary.LittleEndian, int32(len(key))); err != nil {
			return err
		}
		// 写入每个 Key 的内容
		if _, err := buf.WriteString(key); err != nil {
			return err
		}
	}

	// 写入 Values 的长度
	if err := binary.Write(buf, binary.LittleEndian, int32(len(columns.Values))); err != nil {
		return err
	}
	for _, value := range columns.Values {
		// 写入每个 Value 的长度
		if err := binary.Write(buf, binary.LittleEndian, int32(len(value))); err != nil {
			return err
		}
		// 写入每个 Value 的内容
		if _, err := buf.Write(value); err != nil {
			return err
		}
	}

	// 写入 Creates
	if err := binary.Write(buf, binary.LittleEndian, int32(len(columns.Creates))); err != nil {
		return err
	}
	for _, create := range columns.Creates {
		if err := binary.Write(buf, binary.LittleEndian, create); err != nil {
			return err
		}
	}

	// 写入 Expires
	if err := binary.Write(buf, binary.LittleEndian, int32(len(columns.Expires))); err != nil {
		return err
	}
	for _, expire := range columns.Expires {
		if err := binary.Write(buf, binary.LittleEndian, expire); err != nil {
			return err
		}
	}

	return nil
}

func (m *SnapshotManager) parallelLoad(data *ColumnData, total int) error {
	//m.cache.Clear()

	var wg sync.WaitGroup
	workers := runtime.NumCPU()
	chunk := total / workers

	for i := 0; i < workers; i++ {
		wg.Add(1)
		start := i * chunk
		end := start + chunk
		if i == workers-1 {
			end = total
		}

		go func(keys []string, values [][]byte, creates, expires []int64) {
			defer wg.Done()
			for i := range keys {
				if expires[i] > 0 && expires[i] < time.Now().Unix() {
					continue
				}
				m.cache.Add(keys[i], byteview.ByteView{B: values[i]})
			}
		}(data.Keys[start:end], data.Values[start:end],
			data.Creates[start:end], data.Expires[start:end])
	}

	wg.Wait()
	return nil
}
