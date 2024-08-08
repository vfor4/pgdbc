package elephas

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
)

type Rows struct {
	cols   []string
	oids   []int
	datas  []any
	reader *Reader
}

// Columns returns the names of the columns. The number of
// columns of the result is inferred from the length of the
// slice. If a particular column name isn't known, an empty
// string should be returned for that entry.
func (r *Rows) Columns() []string {
	return r.cols
}

// Close closes the rows iterator.
func (r *Rows) Close() error {
	panic("not implemented")
}

// Next is called to populate the next row of data into
// the provided slice. The provided slice will be the same
// size as the Columns() are wide.
//
// Next should return io.EOF when there are no more rows.
//
// The dest should not be written to outside of Next. Care
// should be taken when closing Rows not to modify
// a buffer held in dest.
func (r *Rows) Next(dest []driver.Value) error {
	r.readDataRow(dest)
	return nil
}

func (r *Rows) readDataRow(dest []driver.Value) error {
	msgType, err := r.reader.ReadByte()
	if err != nil {
		return fmt.Errorf("Expected msgType DataRow('D') but got msgType %v", msgType)
	}
	_, err = r.reader.ReadBytesToUint32(4)
	if err != nil {
		return errors.New("readDataRow: Failed to read msgLen")
	}
	fieldCount, err := r.reader.ReadBytesToUint16(2)
	if err != nil {
		return errors.New("readDataRow: Failed to read fieldCount")
	}
	for i := 0; i < int(fieldCount); i++ {
		colLen, err := r.reader.ReadBytesToUint32(4)
		if err != nil {
			return errors.New("readDataRow: Failed to read colLen")
		}
		data, err := r.reader.ReadBytesToAny(colLen, r.oids[i])
		if err != nil {
			return errors.New("readDataRow: Failed to read data")
		}
		dest[i] = data
	}
	log.Println(dest)
	return nil
}
