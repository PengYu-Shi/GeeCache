package byteview

type ByteView struct {
	B []byte
}

func (v ByteView) Len() int {
	return len(v.B)
}

func (v ByteView) ByteSlice() []byte {
	return CloneBytes(v.B)
}

func (v ByteView) String() string {
	return string(v.B)
}

func CloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
