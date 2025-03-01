package snapshot

import "time"

type Meta struct {
	Version     int       `json:"version"`
	Created     time.Time `json:"created"`
	ItemCount   int       `json:"itemCount"`
	DataHash    string    `json:"dataHash"`
	Compression string    `json:"compression"`
}

type ColumnData struct {
	Keys    []string `binary:"lenprefix=uint16"` // 长度前缀为2字节
	Values  [][]byte `binary:"lenprefix=uint32"` // 长度前缀为4字节
	Creates []int64
	Expires []int64
}
