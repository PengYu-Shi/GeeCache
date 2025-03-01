package snapshot

import (
	"GeeCacheNode/gee/byteview"
	"GeeCacheNode/gee/lru"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func filterAndSort(items []*lru.Content) []*lru.Content {
	now := time.Now().Unix()
	filtered := make([]*lru.Content, 0, len(items))

	for _, item := range items {
		if item.ExpiredTime == 0 || item.ExpiredTime > now {
			filtered = append(filtered, item)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreateAt < filtered[j].CreateAt
	})

	return filtered
}

func atomicSwitch(tmpPath string) error {
	tmpLink := filepath.Join(snapshotBase, "current.tmp")
	os.Remove(tmpLink)
	if err := os.Symlink(tmpPath, tmpLink); err != nil {
		log.Println("[Snap Manager] utils.atomicSwitch error")
		return err
	}
	return os.Rename(tmpLink, currentLink)
}

func computeHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}

	// 添加文件大小作为校验因子
	info, _ := f.Stat()
	binary.Write(h, binary.LittleEndian, info.Size())

	return fmt.Sprintf("%x", h.Sum(nil))
}

func saveMeta(meta Meta, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(meta)
}

func toColumns(items []*lru.Content) *ColumnData {
	data := &ColumnData{
		Keys:    make([]string, len(items)),
		Values:  make([][]byte, len(items)),
		Creates: make([]int64, len(items)),
		Expires: make([]int64, len(items)),
	}

	for i, item := range items {
		value := item.Value.(byteview.ByteView)
		data.Keys[i] = item.Key
		data.Values[i] = value.ByteSlice()
		data.Creates[i] = item.CreateAt
		data.Expires[i] = item.ExpiredTime
	}
	return data
}

func validate(metaPath, dataPath string) (Meta, error) {
	var meta Meta

	// 1. 读取并解析元数据文件
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return meta, fmt.Errorf("打开元数据文件失败: %w", err)
	}
	defer metaFile.Close()

	if err := json.NewDecoder(metaFile).Decode(&meta); err != nil {
		return meta, fmt.Errorf("解析元数据失败: %w", err)
	}

	// 2. 校验数据文件哈希
	actualHash := computeHash(dataPath)
	if actualHash != meta.DataHash {
		return meta, fmt.Errorf("数据文件校验失败，期望哈希:%s 实际哈希:%s",
			meta.DataHash, actualHash)
	}

	// 3. 检查快照版本兼容性
	if meta.Version != 1 {
		return meta, fmt.Errorf("不支持的快照版本: %d", meta.Version)
	}

	return meta, nil
}
