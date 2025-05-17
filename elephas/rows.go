package elephas

import (
	"database/sql/driver"
	"fmt"
)

type Row struct {
	cols       []string
	oids       []uint32
	colFormats []uint16
	reader     *Reader
}

func (r *Row) Columns() []string {
	return r.cols
}

func (r *Row) Close() error {
	return nil
}

// int64
// float64
// bool
// []byte
// string
// time.Time
func (r *Row) Next(dest []driver.Value) error {
	return ReadDataRow(dest, r)
}

func ReadDataRow(dest []driver.Value, r *Row) error {
	msgType, err := r.reader.ReadByte()
	if err != nil {
		panic(err)
	}
	if msgType == commandComplete {
		_, err = ReadCommandComplete(r.reader)
		return err
	}
	if msgType != dataRow {
		panic(fmt.Errorf("expected data row - D(68) type but got %v", msgType))
	}
	if _, err = r.reader.Read4Bytes(); err != nil {
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
