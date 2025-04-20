package elephas

import (
	"fmt"
	"log"
)

type ErrorLevel string

const (
	ERROR ErrorLevel = "ERROR"
	FATAL ErrorLevel = "FATAL"
	// PANIC ErrorLevel = "PANIC"
	// WARNING ErrorLevel = "WARNING"
)

type ErrorResponse error

func ReadErrorResponse(r *Reader) (ErrorResponse, error) {
	var severity string
	_, err := r.Read4Bytes()
	if err != nil {
		return nil, err
	}
	for {
		field, err := r.Reader.ReadByte()
		if err != nil {
			return nil, err
		}
		switch field {
		case 'S':
			s, err := r.Reader.ReadBytes(0)
			s = s[:len(s)-1]
			if err != nil {
				return nil, err
			}
			severity = string(s)
		case 'M':
			m, err := r.Reader.ReadBytes(0)
			if err != nil {
				return nil, err
			}
			em := string(m)
			if severity == string(FATAL) {
				log.Fatal(em)
			}
			return fmt.Errorf(em), nil
		default:
			r.Reader.ReadBytes(0)
		}
	}
}
