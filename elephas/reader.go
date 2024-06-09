package elephas

import (
	"bufio"
	"encoding/binary"
	"io"
)

type Reader struct {
	*bufio.Reader
}

func NewReader(r *bufio.Reader) *Reader {
	return &Reader{r}
}

func (r Reader) ReadManyBytes(size uint) (uint32, error) {
	b := make([]byte, size)
	_, err := io.ReadFull(r, b)
	return binary.BigEndian.Uint32(b), err
}
