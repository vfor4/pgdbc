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

func (b *Buffer) buildStartUpMsg() []byte {
	b.Write([]byte{0, 0, 0, 0}) // placeholder for length of message contents
	b.Write((binary.BigEndian.AppendUint32([]byte{}, 196608)))
	b.WriteString("user")
	b.WriteByte(0)
	b.WriteString("postgres")
	b.WriteByte(0)
	b.WriteString("database")
	b.WriteByte(0)
	b.WriteString("record")
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
	binary.BigEndian.PutUint32(data[1:], uint32(len(data)-1)) // dont count letter 'p'
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
	finalQuery := query
	for _, arg := range args {
		finalQuery = strings.Replace(finalQuery, "?", aToString(arg.Value), 1)
	}
	log.Println(finalQuery)
	b.WriteByte('Q')
	initLen := []byte{0, 0, 0, 0}
	binary.BigEndian.PutUint32(initLen, uint32(len(finalQuery)+5)) //4: the length itself; 1:the c-string ending
	b.Write(initLen)
	b.WriteString(finalQuery)
	b.WriteByte(0)
	data := b.Bytes()
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
