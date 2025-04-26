package elephas

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log"
)

type Rows struct {
	cols       []string
	oids       []uint32
	colFormats []uint16
	reader     *Reader
}

func (r *Rows) Columns() []string {
	return r.cols
}

func (r *Rows) Close() error {
	for {
		t, err := r.reader.ReadByte()
		if err != nil {
			return err
		}
		switch t {
		case commandComplete:
			_ = ReadCommandComplete(r)
		case readyForQuery:
			return ReadReadyForQuery(r)
		default:
			panic(errors.New("Close should be here"))
		}
	}
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
func ReadReadyForQuery(r *Rows) error {
	_, err := r.reader.Read4Bytes()
	if err != nil {
		return err
	}
	s, _ := r.reader.ReadByte()
	if s == 'I' {
		log.Print("Status is Idle")
	}
	return nil
}
func ReadCommandComplete(r *Rows) error {
	l, err := r.reader.Read4Bytes()
	if err != nil {
		return err
	}
	_, _ = r.reader.Discard(int(l - 4))
	return io.EOF
}

func ReadDataRow(dest []driver.Value, r *Rows) error {
	msgType, err := r.reader.ReadByte()
	if err != nil {
		panic(err)
	}
	if msgType == commandComplete {
		return ReadCommandComplete(r)
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
