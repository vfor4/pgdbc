package elephas

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
)

type Rows struct {
	cols   []string
	oids   []int
	datas  []any
	reader *Reader
	conn   net.Conn
}

func (r *Rows) Columns() []string {
	return r.cols
}

func (r *Rows) Close() error {
	r.reader.Discard(r.reader.Buffered())
	return nil
}

func (r *Rows) Next(dest []driver.Value) error {
	return ReadDataRow(dest, r)
}

func ReadDataRow(dest []driver.Value, r *Rows) error {
	msgType, err := r.reader.ReadByte()
	if err != nil {
		return fmt.Errorf("readDataRow: Failed to read msgType")
	}
	if msgType == commandComlete {
		return io.EOF
	}
	if msgType != dataRow {
		return fmt.Errorf("Expected msgType DataRow('D') but got msgType %v", msgType)
	}
	_, err = r.reader.ReadBytesToUint32()
	if err != nil {
		return errors.New("readDataRow: Failed to read msgLen")
	}
	fieldCount, err := r.reader.ReadBytesToUint16()
	if err != nil {
		return errors.New("readDataRow: Failed to read fieldCount")
	}
	for i := range int(fieldCount) {
		colLen, err := r.reader.ReadBytesToUint32()
		if err != nil {
			return errors.New("readDataRow: Failed to read colLen")
		}
		data, err := r.reader.ReadBytesToAny(colLen, r.oids[i])
		if err != nil {
			return fmt.Errorf("ReadBytesToAny error: %v", err)
		}
		dest[i] = data
	}
	return nil
}
