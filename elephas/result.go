package elephas

import (
	"fmt"
	"strconv"
	"strings"
)

type Result struct {
	reader *Reader
}

// LastInsertId returns the database's auto-generated ID
// after, for example, an INSERT into a table with primary
// key.
func (re Result) LastInsertId() (int64, error) {
	panic("todo")
}

// RowsAffected returns the number of rows affected by the
// query
func (re Result) RowsAffected() (int64, error) {
	if t, err := re.reader.Reader.ReadByte(); err != nil {
		return 0, fmt.Errorf("read type faild: %w", err)
	} else if t != commandComlete {
		return 0, fmt.Errorf("Expected command complete but got: %v", t)
	}
	l, err := re.reader.ReadBytesToUint32()
	if err != nil {
		return 0, fmt.Errorf("failed to read length, %w", err)
	}
	tag := make([]byte, l-4-1) // - length - null terminated
	_, err = re.reader.Read(tag)
	if err != nil {
		return 0, fmt.Errorf("Failed to read tag, %w", err)
	}
	tags := strings.Split(string(tag), " ")
	rows, err := strconv.Atoi(tags[1])
	if err != nil {
		return 0, fmt.Errorf("Atoi failed to convert tag, %w", err)
	}
	return int64(rows), nil
}
