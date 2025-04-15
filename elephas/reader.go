package elephas

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

type Reader struct {
	*bufio.Reader
}

type DataRow struct {
}

func NewReader(r *bufio.Reader) *Reader {
	return &Reader{r}
}

func (r Reader) ReadBytesToUint32() (uint32, error) {
	b := make([]byte, 4)
	_, err := io.ReadFull(r, b)
	return binary.BigEndian.Uint32(b), err
}

func (r Reader) ReadBytesToUint16() (uint16, error) {
	b := make([]byte, 2)
	_, err := io.ReadFull(r, b)
	return binary.BigEndian.Uint16(b), err
}

func (r Reader) ReadCommandComplete() (string, error) {
	if t, err := r.ReadByte(); err != nil {
		return "", errors.New("unable to read msg type")
	} else if t != commandComlete {
		return "", fmt.Errorf("expect msg type is commandComplete but got (%v)", t)
	}
	_, err := r.ReadBytesToUint32()
	if err != nil {
		return "", err
	}
	cmdTag, err := r.ReadString(0)
	return strings.Trim(cmdTag, "\x00"), nil
}

func (r Reader) ReadReadyForQuery() (TransactionStatus, error) {
	if t, err := r.ReadByte(); err != nil {
		return E, errors.New("unable to read msg type")
	} else if t != readyForQuery {
		return E, fmt.Errorf("expect msg type is readForQuery but got (%v)", t)
	}
	_, err := r.ReadBytesToUint32()
	if err != nil {
		return E, err
	}
	txStatus, err := r.ReadByte()
	if err != nil {
		return E, err
	}
	return TransactionStatus(txStatus), nil
}

func (r Reader) ReadBytesToAny(size uint32, dataType int) (any, error) {
	b := make([]byte, size)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	switch dataType {
	case 23:
		v, err := strconv.Atoi(string(b))
		if err != nil {
			return nil, nil
		}
		return v, nil
	case 25, 1043:
		return string(b), nil
	case 16:
		return strconv.ParseBool(string(b))
	case 1114:
		return time.Parse("2006-01-02 15:04:05.000000", string(b))
	case 20:
		bigBuf := make([]byte, 8)
		bigBuf = append(bigBuf, b...)
		// bigint
		return int64(binary.BigEndian.Uint64(bigBuf)), nil
	default:
		//select oid, typname from pg_type where oid = ?;
		panic(fmt.Sprintf("the OID type %v is not implemented", dataType))
	}
}

func (r Reader) handleAuthResp(authType uint32) ([]byte, error) {
	if t, err := r.Reader.ReadByte(); err != nil {
		return nil, err
	} else if t != authMsgType {
		return nil, fmt.Errorf("expect message type is authentication (%v) but got: %v", authMsgType, t)
	}
	l, err := r.ReadBytesToUint32()
	l -= 8 //
	if err != nil {
		return nil, err
	}
	respAuthType, err := r.ReadBytesToUint32()
	if respAuthType != authType {
		return nil, fmt.Errorf("expect authentication type (%v) but got: %v", authType, respAuthType)
	}
	if l == 0 { // the end of the response
		return nil, nil
	}
	d := make([]byte, l)
	if _, err := io.ReadFull(r.Reader, d); err != nil {
		return nil, err
	}
	return d, nil
}

func ReadSimpleQueryRes(r *Reader) (Rows, error) {
	msgType, err := r.ReadByte()
	if err != nil {
		return Rows{}, err
	}
outer:
	for {
		switch msgType {
		case errorResponseMsg:
			var severity string
			_, err := r.ReadBytesToUint32()
			if err != nil {
				return Rows{}, err
			}
			for {
				field, err := r.Reader.ReadByte()
				if err != nil {
					return Rows{}, err
				}
				switch field {
				case 'S':
					s, err := r.Reader.ReadBytes(0)
					s = s[:len(s)-1]
					if err != nil {
						return Rows{}, err
					}
					severity = string(s)
				case 'M':
					errMsg, err := r.Reader.ReadBytes(0)
					if err != nil {
						return Rows{}, err
					}
					if severity == "FATAL" {
						log.Fatal(string(errMsg))
					} else {
						log.Println(string(errMsg))
						// panic(string(errMsg))
					}
					break outer
				default:
					// TODO
					r.Reader.ReadBytes(0)
				}
			}
		case rowDescription:
			break outer
		default:
			panic(fmt.Sprintf("Not expected type %v", msgType))
		}
	}
	_, err = r.ReadBytesToUint32()
	if err != nil {
		return Rows{}, errors.New("readRowDescription: Failed to read msgLen")
	}
	fieldCount, err := r.ReadBytesToUint16()
	if err != nil {
		return Rows{}, errors.New("readRowDescription: Failed to read fieldCount")
	}
	var rows Rows
	for range int(fieldCount) {
		fieldName, err := r.ReadString(0)
		if err != nil {
			return Rows{}, errors.New("readRowDescription: Failed to read fieldName")
		}
		rows.cols = append(rows.cols, fieldName)
		r.Discard(4 + 2)
		oid, err := r.ReadBytesToUint32()
		if err != nil {
			return Rows{}, errors.New("readRowDescription: Failed to read oid")
		}
		rows.oids = append(rows.oids, int(oid))
		r.Discard(2 + 4 + 2)
	}
	rows.reader = r
	return rows, nil
}
