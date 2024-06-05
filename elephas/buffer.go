package elephas

import (
	"bytes"
	"encoding/binary"
)

type Buffer struct {
	data bytes.Buffer
}

func (b *Buffer) WriteInt32(n int32) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(n))
	b.data.Write(buf)
}
func (b *Buffer) WriteBytes(data []byte) {
	b.data.Write(data)
}
func (b *Buffer) WriteString(str string) {
	b.data.WriteString(str)
	b.data.Write([]byte{','})
}
func (b *Buffer) CalculateSize(prefix int) {
	l := b.data.Len()
	data := b.data.Bytes()

	binary.BigEndian.PutUint32(data[prefix:], uint32(l-prefix))

	b.data.Reset()
	b.data.Write(data)
}
func (b *Buffer) Data() []byte {
	return b.data.Bytes()
}
