package elephas

import (
	"bufio"
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type Reader struct {
	*bufio.Reader
}

func NewReader(r *bufio.Reader) *Reader {
	return &Reader{r}
}

func (r Reader) Read4Bytes() (int32, error) {
	b := make([]byte, 4)
	_, err := io.ReadFull(r, b)
	return int32(binary.BigEndian.Uint32(b)), err
}

func (r Reader) Read4BytesUint32() (uint32, error) {
	b := make([]byte, 4)
	_, err := io.ReadFull(r, b)
	return binary.BigEndian.Uint32(b), err
}

func (r Reader) Read2Bytes() (uint16, error) {
	b := make([]byte, 2)
	_, err := io.ReadFull(r, b)
	return binary.BigEndian.Uint16(b), err
}

func CheckReadyForQuery(r *Reader, txs TransactionStatus) error {
	if r.Buffered() > 0 {
		if t, err := r.ReadByte(); err != nil {
			return err
		} else if t != readyForQuery {
			return fmt.Errorf("Expected ReadForQuery but got (%v)", t)
		}
		_, err := r.Read4Bytes()
		if err != nil {
			return err
		}
		if s, err := r.ReadByte(); err != nil {
			return err
		} else if TransactionStatus(s) == E {
			return fmt.Errorf("Expected %v status but got ERROR(%v)", txs, E)
		}
	}
	return nil
}

func (r Reader) ReadBytesToAny(size int32, oid uint32, format uint16) (any, error) {
	if size == NULL_SIZE {
		// select null
		return nil, nil
	}
	b := make([]byte, size)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	switch format {
	case uint16(fmtText):
		switch oid {
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
			bigInt, err := strconv.ParseInt(string(b), 10, 32)
			if err != nil {
				panic(err)
			}
			return bigInt, nil

		default:
			//select oid, typname from pg_type where oid = ?;
			panic(fmt.Sprintf("the OID type %v is not implemented", oid))
		}
	case uint16(fmtBinary):
		switch size {
		case 2:
			return binary.BigEndian.Uint16(b), nil
		case 4:
			return binary.BigEndian.Uint32(b), nil
		case 8:
			return binary.BigEndian.Uint64(b), nil
		default:
			panic("todo")
		}
	default:
		return nil, fmt.Errorf("unexpected format type")
	}
}

func (r Reader) handleAuthResp(authType uint32) ([]byte, error) {
	if t, err := r.Reader.ReadByte(); err != nil {
		return nil, err
	} else if t != authMsgType {
		return nil, fmt.Errorf("expect message type is authentication (%v) but got: %v", authMsgType, t)
	}
	l, err := r.Read4Bytes()
	l -= 8 //
	if err != nil {
		return nil, err
	}
	respAuthType, err := r.Read4BytesUint32()
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

func ReadRows(r *Reader) (Row, error) {
	msgType, err := r.ReadByte()
	if err != nil {
		panic(err)
	}
	switch msgType {
	case rowDescription:
		_, err := r.Read4Bytes() // msgLen
		if err != nil {
			panic(err)
		}
		fieldCount, err := r.Read2Bytes()
		if err != nil {
			panic(err)
		}
		var rows Row
		for range fieldCount {
			fieldName, err := r.ReadString(0)
			if err != nil {
				return Row{}, errors.New("readRowDescription: Failed to read fieldName")
			}
			rows.cols = append(rows.cols, fieldName)
			r.Discard(4 + 2) //skip tableOid, column index

			typeOid, err := r.Read4BytesUint32()
			if err != nil {
				panic(err)
			}
			rows.oids = append(rows.oids, typeOid)
			r.Discard(2 + 4) // skip column length, type modifier

			fmt, err := r.Read2Bytes()
			if err != nil {
				panic(err)
			}
			rows.colFormats = append(rows.colFormats, fmt)

		}
		rows.reader = r
		return rows, nil
	case errorResponseMsg:
		return Row{}, ReadErrorResponse(r)
	default:
		return Row{}, fmt.Errorf("Not expected type %v", msgType)
	}
}

func ReadStmtComplete(r *Reader, c byte) error {
	t, err := r.ReadByte()
	if err != nil {
		return err
	}
	if t != c {
		return fmt.Errorf("Expected Complete with %v but got %v", c, t)
	}
	if n, _ := r.Discard(4); n != 4 {
		return errors.New("Failed to discard Complete command type")
	}
	return nil
}

func ReadResult(r *Reader, q string) (driver.Result, error) {
	commands := len(strings.SplitAfter(strings.TrimRight(q, ";"), ";"))
	n := 0
	for range commands {
		n++
		t, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		switch t {
		case errorResponseMsg:
			return nil, ReadErrorResponse(r)
		case commandComplete:
			_, err := r.Read4Bytes()
			if err != nil {
				return nil, err
			}
			tag, err := r.ReadString(0)
			if err != nil {
				return nil, err
			}
			if n == commands {
				if strings.HasPrefix(tag, "CREATE") {
					return driver.ResultNoRows, nil
				}
				return driver.RowsAffected(0), nil
			}
		default:
			panic(t)
		}
	}
	return nil, nil
}
