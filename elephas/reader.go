package elephas

import (
	"bufio"
	"io"
)

type Reader struct {
	*bufio.Reader
}

func NewReader(r *bufio.Reader) *Reader {
	return &Reader{r}
}
func (r Reader) ReadMessageType() (byte, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return b, nil
}

func (r Reader) ReadManyBytes(size uint) ([]byte, error) {
	b := make([]byte, size)
	_, err := io.ReadFull(r, b)
	return b, err
}
