package elephas

import "fmt"

type ErrorLevel string

const (
	ERROR ErrorLevel = "ERROR"
	FATAL ErrorLevel = "FATAL"
	// PANIC ErrorLevel = "PANIC"
	// WARNING ErrorLevel = "WARNING"
)

type ErrorResponse struct {
	severity ErrorLevel
	message  string
	detail   string
}

func (e ErrorResponse) Error() string {
	em := fmt.Sprintf("%s: %s", e.severity, e.message)
	if e.detail != "" {
		return fmt.Sprintf("%s: %s", em, e.detail)
	}
	return em
}

func ReadErrorResponse(r *Reader) ErrorResponse {
	var er ErrorResponse
	_, err := r.Read4Bytes() // length
	if err != nil {
		panic(err)
	}
	for {
		fieldType, err := r.Reader.ReadByte()
		if err != nil {
			panic(err)
		}
		switch fieldType {
		case 'S':
			s, err := r.Reader.ReadBytes(0)
			if err != nil {
				panic(err)
			}
			er.severity = ErrorLevel(s)
		case 'V', 'C':
			_, _ = r.Reader.ReadBytes(0)
		case 'M':
			s, err := r.Reader.ReadString(0)
			if err != nil {
				panic(err)
			}
			er.message = string(s)
		case 'D':
			s, err := r.Reader.ReadString(0)
			if err != nil {
				panic(err)
			}
			er.detail = s
		default:
			r.Reader.Discard(r.Reader.Buffered())
			return er
		}
	}
}
