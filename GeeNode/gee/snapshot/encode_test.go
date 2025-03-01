package snapshot

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	// 构造测试数据
	orig := &ColumnData{
		Keys:    []string{"key1", "key2", "key3", "key4", "key5", "key6", "key7"},
		Values:  [][]byte{[]byte("value1"), []byte("value2"), []byte("value2"), []byte("value2"), []byte("value2"), []byte("value2"), []byte("value2")},
		Creates: []int64{1620000000, 1620000001, 1620000001, 1620000001, 1620000001, 1620000001, 1620000001},
		Expires: []int64{1620003600, 1620003601, 1620003601, 1620003601, 1620003601, 1620003601, 1620003601},
	}

	// 序列化
	data, err := orig.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	// 反序列化
	decoded := new(ColumnData)
	if err := decoded.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	// 验证数据一致性
	if !reflect.DeepEqual(orig, decoded) {
		t.Errorf("数据不一致\n原始:%+v\n解码:%+v", orig, decoded)
	}
}
func TestDecodeMultipleKeys(t *testing.T) {
	// 构造测试数据
	orig := &ColumnData{
		Keys:    []string{"key1", "key2", strings.Repeat("a", 5000)}, // 长key测试
		Values:  [][]byte{[]byte("value1"), []byte("value2"), make([]byte, 10240)},
		Creates: []int64{1, 2, 3},
		Expires: []int64{100, 200, 300},
	}

	// 编码
	encoder := NewEncoder()
	compressed, err := encoder.Encode(orig)
	assert.NoError(t, err)

	// 解码
	decoder := NewDecoder()
	decoded, err := decoder.Decode(compressed)
	assert.NoError(t, err)
	defer decoder.Close()

	// 验证
	assert.Equal(t, orig.Keys, decoded.Keys)
	assert.Equal(t, orig.Values, decoded.Values)
	assert.Equal(t, orig.Creates, decoded.Creates)
	assert.Equal(t, orig.Expires, decoded.Expires)

	// 验证数据边界
	assert.Len(t, decoded.Keys[2], 5000)
	assert.Len(t, decoded.Values[2], 10240)
}
