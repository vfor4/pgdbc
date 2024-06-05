package elephas

import (
	"bufio"
	"context"
	"database/sql/driver"
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

// https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-START-UP
func (c *Connection) makeHandShake(ctx context.Context) error {
	var b Buffer
	// size of data (4 byte) without first byte (message type)
	b.WriteBytes([]byte{0, 0, 0, 0})
	b.WriteInt32(196608)
	b.WriteString("user")
	b.WriteString(c.cfg.User)
	b.WriteString("database")
	b.WriteString(c.cfg.Database)
	b.WriteBytes([]byte{0})

	b.CalculateSize(0)
	log.Print(b.Data())
	if _, err := c.conn.Write(b.Data()); err != nil {
		return err
	}
	msgType, err := c.reader.ReadMessageType()
	if err != nil {
		return err
	}
	log.Print(string(msgType))
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
