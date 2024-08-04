package elephas

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

type Reader struct {
	*bufio.Reader
}

func NewReader(r *bufio.Reader) *Reader {
	return &Reader{r}
}

func (r Reader) ReadBytesToUint32(size uint) (uint32, error) {
	b := make([]byte, size)
	_, err := io.ReadFull(r, b)
	return binary.BigEndian.Uint32(b), err
}

func (r Reader) ReadBytesToUint16(size uint) (uint16, error) {
	b := make([]byte, size)
	_, err := io.ReadFull(r, b)
	return binary.BigEndian.Uint16(b), err
}

func (r Reader) handleAuthResp(authType uint32) ([]byte, error) {
	if t, err := r.Reader.ReadByte(); err != nil {
		return nil, err
	} else if t != authMsgType {
		return nil, fmt.Errorf("expect message type is authentication (%v) but got: %v", authMsgType, t)
	}
	l, err := r.ReadBytesToUint32(4)
	l -= 8 //
	if err != nil {
		return nil, err
	}
	respAuthType, err := r.ReadBytesToUint32(4)
	if respAuthType != authType {
		return nil, fmt.Errorf("expect authentication type (%v) but got: %v", authType, respAuthType)
	}
	log.Println("respAuthType", respAuthType)
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

func (r Reader) readRowDescription() (Rows, error) {
	msgType, err := r.ReadByte()
	if err != nil {
		return Rows{}, err
	}
	if msgType != rowDescription {
		return Rows{}, fmt.Errorf("Expect Row Description type but got %v", msgType)
	}
	msgLen, err := r.ReadBytesToUint32(4)
	if err != nil {
		return Rows{}, errors.New("readRowDescription: Failed to read msgLen")
	}
	fieldCount, err := r.ReadBytesToUint16(2)
	if err != nil {
		return Rows{}, errors.New("readRowDescription: Failed to read fieldCount")
	}
	log.Println(msgLen, fieldCount)
	log.Println(r.ReadString(0))
	return Rows{}, nil
}
