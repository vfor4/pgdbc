package elephas

import (
	"errors"
	"fmt"
	"log"
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
	for {
		t, err := re.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		_, err = re.reader.Read4Bytes()
		switch t {
		case rowDescription:
			fieldCount, err := re.reader.Read2Bytes()
			if err != nil {
				return 0, err
			}
			if fieldCount != 1 {
				return 0, fmt.Errorf("Field count is larger than 1, actual:%v", fieldCount)
			}
		case commandComlete:
			if err != nil {
				return 0, err
			}
			b, err := re.reader.Reader.ReadBytes(0)
			if err != nil {
				return 0, err
			}
			tags := strings.Split(string(b), " ")
			log.Println(tags)
			return 0, nil
		case errorResponseMsg:
			var severity string
			_, err := re.reader.Read4Bytes()
			if err != nil {
				return 0, err
			}
			for {
				field, err := re.reader.ReadByte()
				if err != nil {
					return 0, err
				}
				switch field {
				case 'S':
					s, err := re.reader.ReadBytes(0)
					s = s[:len(s)-1]
					if err != nil {
						return 0, err
					}
					severity = string(s)
				case 'M':
					errMsg, err := re.reader.ReadBytes(0)
					if err != nil {
						return 0, err
					}
					if severity == "FATAL" {
						log.Fatal(string(errMsg))
					} else {
						log.Println(string(errMsg))
					}
					return 0, errors.New(string(errMsg))
				default:
					// TODO
					re.reader.ReadBytes(0)
				}
			}
		default:
			panic(string(t))
		}

	}
}

// RowsAffected returns the number of rows affected by the
// query
func (re Result) RowsAffected() (int64, error) {
	if t, err := re.reader.Reader.ReadByte(); err != nil {
		return 0, fmt.Errorf("read type faild: %w", err)
	} else if t != commandComlete {
		return 0, fmt.Errorf("Expected command complete but got: %v", t)
	}
	l, err := re.reader.Read4Bytes()
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
