package elephas

import (
	"bufio"
	"encoding/binary"
	"fmt"
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

func (r Reader) handleAuthResp(authType uint32) ([]byte, error) {
	if t, err := r.Reader.ReadByte(); err != nil {
		return nil, err
	} else if t != authMsgType {
		return nil, fmt.Errorf("expect message type is authentication (%v) but got: %v", authMsgType, t)
	}
	l, err := r.ReadManyBytes(4)
	l -= 8 //
	if err != nil {
		return nil, err
	}
	respAuthType, err := r.ReadManyBytes(4)
	if respAuthType != authType {
		return nil, fmt.Errorf("expect authentication type (%v) but got: %v", authType, respAuthType)
	}
	if l == 0 { // the end of the response
		return nil, nil
	}
	// i like those letters 't', 'l'. They confuse the reader, haha
	d := make([]byte, l)
	if _, err := io.ReadFull(r.Reader, d); err != nil {
		return nil, err
	}
	return d, nil
}
