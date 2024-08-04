package elephas

import "database/sql/driver"

type Rows struct{}

// Columns returns the names of the columns. The number of
// columns of the result is inferred from the length of the
// slice. If a particular column name isn't known, an empty
// string should be returned for that entry.
func (r Rows) Columns() []string {
	panic("not implemented")
}

// Close closes the rows iterator.
func (r Rows) Close() error {
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
func (r Rows) Next(dest []driver.Value) error {
	panic("not implemented")
}
