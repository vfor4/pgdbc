package elephas

import (
	"bufio"
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/binary"
	"log"
	"net"
	"time"
)

// Conn
type Connection struct {
	cfg    *Config
	conn   net.Conn
	reader *Reader
}

func (c *Connection) Prepare(query string) (driver.Stmt, error) {
	return nil, nil
}

func (c *Connection) Close() error {
	return nil
}

func (c *Connection) Begin() (driver.Tx, error) {
	return nil, nil
}

func buildStartUpMsg() []byte {
	var b bytes.Buffer
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
	var data []byte
	data = binary.BigEndian.AppendUint32(data, uint32(len(b.Bytes())+4))
	data = append(data, b.Bytes()...)
	return data
}

// https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-START-UP
func (c *Connection) makeHandShake(ctx context.Context) error {
	if _, err := c.conn.Write(buildStartUpMsg()); err != nil {
		log.Fatalf("Failed to make hande shake: %v", err)
		return err
	}

	msgType, err := c.reader.ReadMessageType()
	if err != nil {
		log.Fatalf("Failed to get ReadMessageType: %v", err)
	}
	log.Println(string(msgType))

	switch msgType {
	case authenticationOKMsg:
		saslType, err := c.reader.ReadManyBytes(4)
		if err != nil {
			log.Printf("Failed to get saslType: %v \n", err)
			return err
		}
		log.Println(saslType)
	}
	return nil
}

func NewConnection(ctx context.Context, cfg *Config) (*Connection, error) {
	d := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 5 * time.Minute,
	}

	dConn, err := d.DialContext(ctx, cfg.Network, cfg.Addr)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
		return nil, err
	}
	defer dConn.Close()

	reader := NewReader(bufio.NewReader(dConn))
	conn := Connection{conn: dConn, reader: reader, cfg: cfg}
	if err := conn.makeHandShake(ctx); err != nil {
		log.Fatalf("Failed to make handle shake: %v", err)
		return nil, err
	}

	return &conn, nil
}
