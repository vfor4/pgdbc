package elephas

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
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

func (r Reader) Read4Bytes() (uint32, error) {
	b := make([]byte, 4)
	_, err := io.ReadFull(r, b)
	return binary.BigEndian.Uint32(b), err
}

func (r Reader) Read2Bytes() (uint16, error) {
	b := make([]byte, 2)
	_, err := io.ReadFull(r, b)
	return binary.BigEndian.Uint16(b), err
}

func (r Reader) ReadCommandComplete() (string, error) {
	if t, err := r.ReadByte(); err != nil {
		return "", errors.New("unable to read msg type")
	} else if t != commandComplete {
		return "", fmt.Errorf("expect msg type is commandComplete but got (%v)", t)
	}
	_, err := r.Read4Bytes()
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
	_, err := r.Read4Bytes()
	if err != nil {
		return E, err
	}
	txStatus, err := r.ReadByte()
	if err != nil {
		return E, err
	}
	return TransactionStatus(txStatus), nil
}

func (r Reader) ReadBytesToAny(size uint32, oid uint32, format uint16) (any, error) {
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
		panic("todo binary format")
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
	respAuthType, err := r.Read4Bytes()
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

func ReadSimpleQueryRes(r *Reader, conn net.Conn) (Rows, error) {
	msgType, err := r.ReadByte()
	if err != nil {
		panic(err)
	}
	if msgType == errorResponseMsg {
		errResponse, err := ReadErrorResponse(r)
		if err != nil {
			panic(err)
		}
		return Rows{}, fmt.Errorf("Server response with an error = %+v\n", errResponse.Error())
	} else {
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
			var rows Rows
			for range int(fieldCount) {
				fieldName, err := r.ReadString(0)
				if err != nil {
					return Rows{}, errors.New("readRowDescription: Failed to read fieldName")
				}
				rows.cols = append(rows.cols, fieldName)
				r.Discard(4 + 2) //skip tableOid, column index

				typeOid, err := r.Read4Bytes()
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
			errResponse, err := ReadErrorResponse(r)
			if err != nil {
				panic(err)
			}
			return Rows{}, errResponse
		default:
			return Rows{}, fmt.Errorf("Not expected type %v", msgType)
		}
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
	return nil
}
