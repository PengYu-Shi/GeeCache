package snapshot

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/klauspost/compress/zstd"
	"io"
)

// ZSTDEncoder 列式数据编码器
type ZSTDEncoder struct {
	encoder *zstd.Encoder
}

func NewEncoder() *ZSTDEncoder {
	enc, _ := zstd.NewWriter(nil,
		zstd.WithEncoderLevel(zstd.SpeedBetterCompression),
		zstd.WithEncoderConcurrency(2),
	)
	return &ZSTDEncoder{encoder: enc}
}

// Encode 将列式数据编码为压缩字节流
func (e *ZSTDEncoder) Encode(data *ColumnData) ([]byte, error) {
	var buf bytes.Buffer

	// 1. 编码元信息
	if err := binary.Write(&buf, binary.LittleEndian, uint32(len(data.Keys))); err != nil {
		return nil, err
	}

	// 2. 编码Keys列
	for _, key := range data.Keys {
		if err := binary.Write(&buf, binary.LittleEndian, uint16(len(key))); err != nil {
			return nil, err
		}
		buf.WriteString(key)
	}

	// 3. 编码Values列
	for _, val := range data.Values {
		if err := binary.Write(&buf, binary.LittleEndian, uint32(len(val))); err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	// 4. 编码时间戳列
	if err := binary.Write(&buf, binary.LittleEndian, data.Creates); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.LittleEndian, data.Expires); err != nil {
		return nil, err
	}

	// 5. ZSTD压缩
	compressed := e.encoder.EncodeAll(buf.Bytes(), nil)
	return compressed, nil
}

// ZSTDDecoder 列式数据解码器
type ZSTDDecoder struct {
	decoder *zstd.Decoder
}

func NewDecoder() *ZSTDDecoder {
	dec, _ := zstd.NewReader(nil)
	return &ZSTDDecoder{decoder: dec}
}

// Decode 将压缩字节流解码为列式数据
func (d *ZSTDDecoder) Decode(compressed []byte) (*ColumnData, error) {
	// 1. ZSTD解压缩
	decompressed, err := d.decoder.DecodeAll(compressed, nil)
	if err != nil {
		return nil, err
	}

	// 2. 读取元信息
	r := bytes.NewReader(decompressed)
	var itemCount uint32
	if err := binary.Read(r, binary.LittleEndian, &itemCount); err != nil {
		return nil, err
	}
	readFull := func(r io.Reader, data []byte) error {
		_, err := io.ReadFull(r, data)
		return err
	}

	data := &ColumnData{
		Keys:    make([]string, itemCount),
		Values:  make([][]byte, itemCount),
		Creates: make([]int64, itemCount),
		Expires: make([]int64, itemCount),
	}

	// 3. 解码Keys列
	for i := 0; i < int(itemCount); i++ {
		var keyLen uint16
		if err := binary.Read(r, binary.LittleEndian, &keyLen); err != nil {
			return nil, fmt.Errorf("读取key长度失败[索引%d]: %w", i, err)
		}

		keyBuf := make([]byte, keyLen)
		if err := readFull(r, keyBuf); err != nil {
			return nil, fmt.Errorf("读取key内容失败[keyLen=%d]: %w", keyLen, err)
		}
		data.Keys[i] = string(keyBuf)
	}

	// 4. 解码Values列
	for i := 0; i < int(itemCount); i++ {
		var valLen uint32
		if err := binary.Read(r, binary.LittleEndian, &valLen); err != nil {
			return nil, fmt.Errorf("读取value长度失败[索引%d]: %w", i, err)
		}

		data.Values[i] = make([]byte, valLen)
		if err := readFull(r, data.Values[i]); err != nil {
			return nil, fmt.Errorf("读取value内容失败[valLen=%d]: %w", valLen, err)
		}
	}

	// 5. 解码时间戳列
	if err := binary.Read(r, binary.LittleEndian, &data.Creates); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &data.Expires); err != nil {
		return nil, err
	}
	if r.Len() != 0 {
		return nil, fmt.Errorf("解码完成后仍有%d字节未处理", r.Len())
	}
	return data, nil
}

// Close 释放解码器资源
func (d *ZSTDDecoder) Close() error {
	d.decoder.Close()
	return nil
}
func (c *ColumnData) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// 1. 序列化Keys
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(c.Keys))); err != nil {
		return nil, err
	}
	for _, key := range c.Keys {
		if err := binary.Write(buf, binary.LittleEndian, uint16(len(key))); err != nil {
			return nil, err
		}
		buf.WriteString(key)
	}

	// 2. 序列化Values
	for _, val := range c.Values {
		if err := binary.Write(buf, binary.LittleEndian, uint32(len(val))); err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	// 3. 序列化时间戳
	if err := binary.Write(buf, binary.LittleEndian, c.Creates); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.LittleEndian, c.Expires); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// 实现binary.Read的自定义反序列化
func (c *ColumnData) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)

	// 1. 读取Keys
	var keyCount uint32
	if err := binary.Read(r, binary.LittleEndian, &keyCount); err != nil {
		return err
	}
	c.Keys = make([]string, keyCount)
	for i := range c.Keys {
		var keyLen uint16
		if err := binary.Read(r, binary.LittleEndian, &keyLen); err != nil {
			return err
		}
		key := make([]byte, keyLen)
		if _, err := r.Read(key); err != nil {
			return err
		}
		c.Keys[i] = string(key)
	}

	// 2. 读取Values
	c.Values = make([][]byte, keyCount)
	for i := range c.Values {
		var valLen uint32
		if err := binary.Read(r, binary.LittleEndian, &valLen); err != nil {
			return err
		}
		c.Values[i] = make([]byte, valLen)
		if _, err := r.Read(c.Values[i]); err != nil {
			return err
		}
	}

	// 3. 读取时间戳
	c.Creates = make([]int64, keyCount)
	if err := binary.Read(r, binary.LittleEndian, &c.Creates); err != nil {
		return err
	}
	c.Expires = make([]int64, keyCount)
	if err := binary.Read(r, binary.LittleEndian, &c.Expires); err != nil {
		return err
	}

	return nil
}
