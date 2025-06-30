package elephas

import (
	"bytes"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"hash"
	"strings"

	"mellium.im/sasl"
)

type Buffer struct {
	bytes.Buffer
}

func (b *Buffer) buildStartUpMsg(user, db string) []byte {
	b.Write([]byte{0, 0, 0, 0}) // placeholder for length of message contents
	b.Write((binary.BigEndian.AppendUint32([]byte{}, 196608)))
	b.WriteString("user")
	b.WriteByte(0)
	b.WriteString(user)
	b.WriteByte(0)
	b.WriteString("database")
	b.WriteByte(0)
	b.WriteString(db)
	b.WriteByte(0)
	b.WriteByte(0) // null-terminated c-style string
	data := b.Bytes()
	binary.BigEndian.PutUint32(data, uint32(len(b.Bytes())))
	b.Reset()
	return data
}

func (b *Buffer) buildSASLInitialResponse(initClientResp []byte) []byte {
	b.WriteByte('p')
	b.Write([]byte{0, 0, 0, 0})
	b.WriteString(sasl.ScramSha256.Name)
	b.WriteByte(0)

	initLen := []byte{0, 0, 0, 0}
	binary.BigEndian.PutUint32(initLen, uint32(len(initClientResp)))

	b.Write(initLen)
	b.Write(initClientResp)
	data := b.Bytes()
	binary.BigEndian.PutUint32(data[1:], uint32(len(data)-1)) //  count letter 'p'
	b.Reset()
	return data
}

func (b *Buffer) buildSASLResponse(saslChallenge []byte) []byte {
	b.WriteByte('p')
	initLen := []byte{0, 0, 0, 0}
	binary.BigEndian.PutUint32(initLen, uint32(len(saslChallenge)+4))
	b.Write(initLen)
	b.Write(saslChallenge)
	data := b.Bytes()
	b.Reset()
	return data
}

func (b *Buffer) buildQuery(query string, args []driver.NamedValue) []byte {
	for _, arg := range args {
		query = strings.Replace(query, "?", aToString(arg.Value), 1)
	}
	b.WriteByte(byte(queryCommand))
	initLen := []byte{0, 0, 0, 0}
	binary.BigEndian.PutUint32(initLen, uint32(len(query)+5)) //4: the length itself; 1:the c-string ending
	b.Write(initLen)
	b.WriteString(query)
	b.WriteByte(0)
	data := b.Bytes()
	b.Reset()
	return data
}

var hw hash.Hash = sha256.New()

// TODO have to refactor
func (b *Buffer) buidParseCmd(query string, params int) []byte {
	_, _ = hw.Write([]byte(query))
	name := fmt.Sprintf("%x", hw.Sum(nil))
	for i := range params {
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", i+1), 1)
	}
	b.WriteByte(parseCommand)
	b.Write([]byte{0, 0, 0, 0})
	b.WriteString(name)
	b.WriteByte(0)
	b.WriteString(query)
	b.WriteByte(0)
	buf := binary.BigEndian.AppendUint16(make([]byte, 0, params), uint16(params))
	b.Write(buf)
	for range params {
		b.Write([]byte{0, 0, 0, 0})
	}
	data := b.Bytes()
	binary.BigEndian.PutUint32(data[1:], uint32(len(data)-1))
	b.Reset()
	return data
}

func (b *Buffer) buildFlushCmd() []byte {
	b.WriteByte(flushCommand)
	b.Write([]byte{0, 0, 0, 4})
	data := b.Bytes()
	b.Reset()
	return data
}
func (b *Buffer) buildBindCmd(args []driver.NamedValue, namedStmt string, portalName string) []byte {
	b.WriteByte(bindCommand)
	b.Write([]byte{0, 0, 0, 0})
	b.WriteString(portalName)
	b.WriteByte(0)

	b.WriteString(namedStmt)
	b.WriteByte(0)

	b.Write([]byte{0, 1}) // number of param format
	b.Write([]byte{0, 1}) // binary

	n := len(args)
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(n))
	b.Write(buf) // number of param value
	for _, v := range args {
		if v.Value == nil {
			binary.Write(b, binary.BigEndian, int32(1))
		} else {
			switch vl := v.Value.(type) {
			case string:
				binary.Write(b, binary.BigEndian, int32(len(vl)))
				b.WriteString(vl)
			case int64:
				binary.Write(b, binary.BigEndian, int32(8))
				binary.Write(b, binary.BigEndian, int64(vl))
			case int32:
				binary.Write(b, binary.BigEndian, int32(4))
				binary.Write(b, binary.BigEndian, int32(vl))
			default:
				fmt.Println(vl)
			}
		}
	}
	binary.Write(b, binary.BigEndian, int16(1)) // result col
	binary.Write(b, binary.BigEndian, int16(1)) // result col format 0:text 1:binary

	data := b.Bytes()
	binary.BigEndian.PutUint32(data[1:], uint32(len(data)-1))
	b.Reset()
	return data
}

func (b *Buffer) buildExecuteCmd(namePortal string) []byte {
	b.WriteByte(executeCommand)
	b.Write([]byte{0, 0, 0, 0})
	b.WriteString(namePortal)
	b.WriteByte(0)
	b.Write([]byte{0, 0, 0, 0})
	data := b.Bytes()
	binary.BigEndian.PutUint32(data[1:], uint32(len(data)-1))
	b.Reset()
	return data
}

func (b *Buffer) buildSync() []byte {
	b.WriteByte(syncCommand)
	b.Write([]byte{0, 0, 0, 0})
	data := b.Bytes()
	binary.BigEndian.PutUint32(data[1:], uint32(len(data)-1))
	b.Reset()
	return data
}

func (b *Buffer) buildDescribe(name string) []byte {
	b.WriteByte(describeCommand)
	b.Write([]byte{0, 0, 0, 0})
	b.WriteByte(byte('P')) // S: statement ; P: portal
	b.WriteString(name)
	b.WriteByte(0)
	data := b.Bytes()
	binary.BigEndian.PutUint32(data[1:], uint32(len(data)-1))
	b.Reset()
	return data
}

func (b *Buffer) buildCopyData(byten []byte) []byte {
	b.WriteByte(copyData)
	t := make([]byte, 4)
	binary.BigEndian.PutUint32(t, uint32(len(byten)+4))
	b.Write(t)
	b.Write(byten)
	data := b.Bytes()
	b.Reset()
	return data
}
func (b *Buffer) buildCopyDone() []byte {
	b.WriteByte(copyDone)
	b.Write([]byte{0, 0, 0, 4})
	data := b.Bytes()
	b.Reset()
	return data
}
func aToString(value driver.Value) string {
	if value == nil {
		return "null"
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Sprintf("%v", value)
	}
	return fmt.Sprintf("'%v'", s)
}
