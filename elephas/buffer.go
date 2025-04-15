package elephas

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"log"
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
	log.Println(query)
	for _, arg := range args {
		query = strings.Replace(query, "?", aToString(arg.Value), 1)
	}
	b.WriteByte(queryCommand)
	initLen := []byte{0, 0, 0, 0}
	binary.BigEndian.PutUint32(initLen, uint32(len(query)+5)) //4: the length itself; 1:the c-string ending
	b.Write(initLen)
	b.WriteString(query)
	b.WriteByte(0)
	data := b.Bytes()
	b.Reset()
	return data
}

// TODO have to refactor
func (b *Buffer) buidParseCmd(query string, name string) []byte {
	b.WriteByte(parseCommand)
	b.Write([]byte{0, 0, 0, 0})
	b.WriteString(name)
	b.WriteByte(0)

	b.WriteString(query)
	b.WriteByte(0)

	b.Write([]byte{0, 1}) // number of params
	paramId := []byte{0, 0, 0, 0}
	binary.BigEndian.PutUint32(paramId, uint32(23))
	b.Write(paramId)
	data := b.Bytes()
	binary.BigEndian.PutUint32(data[1:], uint32(len(data)-1))
	b.Reset()
	return data
}

func (b *Buffer) buildFlushCmd() []byte {
	b.WriteByte(flushCommand)
	b.Write([]byte{0, 0, 0, 0})
	data := b.Bytes()
	binary.BigEndian.PutUint32(data[1:], uint32(len(data)-1))
	b.Reset()
	return data
}
func (b *Buffer) buildBindCmd(nameStmt string, portalName string) []byte {
	b.WriteByte(bindCommand)
	b.Write([]byte{0, 0, 0, 0})
	b.WriteString(portalName)
	b.WriteByte(0)

	b.WriteString(nameStmt)
	b.WriteByte(0)

	b.Write([]byte{0, 1}) // number of param format
	b.Write([]byte{0, 1}) // param format code 0: text 1: binary

	b.Write([]byte{0, 1}) // number of param value
	binary.Write(b, binary.BigEndian, int32(4))
	binary.Write(b, binary.BigEndian, int32(5))

	binary.Write(b, binary.BigEndian, int16(1)) // result col
	binary.Write(b, binary.BigEndian, int16(0)) // result col format 0:text 1:binary

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
	b.Write([]byte{0, 0, 0, 0}) // no limit on row
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

func aToString(value driver.Value) string {
	s, ok := value.(string)
	if !ok {
		return fmt.Sprintf("%v", value)
	}
	return fmt.Sprintf("'%v'", s)
}
