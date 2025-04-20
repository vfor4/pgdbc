package elephas

import (
	"database/sql/driver"
	"fmt"
	"io"
	"net"
)

type Rows struct {
	cols       []string
	oids       []uint32
	colFormats []uint16
	reader     *Reader
	conn       net.Conn
}

func (r *Rows) Columns() []string {
	return r.cols
}

func (r *Rows) Close() error {
	return nil
}

// int64
// float64
// bool
// []byte
// string
// time.Time
func (r *Rows) Next(dest []driver.Value) error {
	return ReadDataRow(dest, r)
}

func ReadDataRow(dest []driver.Value, r *Rows) error {
	msgType, err := r.reader.ReadByte()
	if err != nil {
		panic(err)
	}
	if msgType == commandComlete {
		return io.EOF
	}
	if msgType != dataRow {
		panic(fmt.Errorf("expected data row - D(69) type but got %v", msgType))
	}
	_, err = r.reader.Read4Bytes() // skip msgLen
	if err != nil {
		panic(err)
	}
	fieldCount, err := r.reader.Read2Bytes()
	if err != nil {
		panic(err)
	}
	for i := range fieldCount {
		colLen, err := r.reader.Read4Bytes()
		if err != nil {
			panic(err)
		}
		data, err := r.reader.ReadBytesToAny(colLen, r.oids[i], r.colFormats[i])
		if err != nil {
			panic(err)
		}
		dest[i] = data
	}
	return nil
}
