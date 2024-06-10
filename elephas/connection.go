package elephas

import (
	"bufio"
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"mellium.im/sasl"
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
	if _, err := c.conn.Write(b.buildStartUpMsg()); err != nil {
		log.Fatalf("Failed to make hande shake: %v", err)
		return err
	}
	data, err := c.reader.handleAuthResp(SASL)
	if err != nil {
		log.Fatalf("Failed to handle AuthenticationSASL; %v", err)
		return err
	}
	switch string(string(data[:len(data)-2])) {
	case sasl.ScramSha256.Name:
		err = c.authSASL(&b)
		if err != nil {
			return err
		}
	default:
		panic("TODO ScramSha256Plus")
	}
	return nil
}

func (c *Connection) authSASL(b *Buffer) error {
	creds := sasl.Credentials(func() (Username []byte, Password []byte, Identity []byte) {
		return []byte(c.cfg.User), []byte(c.cfg.Password), []byte{}
	})
	client := sasl.NewClient(sasl.ScramSha256, creds)
	_, resp, err := client.Step(nil) // n,,n=postgres,r= nonce
	if err != nil {
		log.Printf("Failed to Step: %v \n", err)
		return err
	}
	_, err = c.conn.Write(b.buildSASLInitialResponse(resp))
	if err != nil {
		log.Printf("Failed to send SASLInitialResponse: %v", err)
		return err
	}
	_, _ = c.reader.ReadByte() // get rid of last byte
	msgType, err := c.reader.ReadByte()
	if msgType == authMsgType { //msgAuthenticationSASLContinue
		msgLen, err := c.reader.ReadManyBytes(4)
		// i smell some duplicate :<
		if err != nil {
			log.Printf("authSASL: Failed to read msgLen: %v \n", err)
			return err
		}
		authType, err := c.reader.ReadManyBytes(4)
		if authType != SASLContinue {
			return fmt.Errorf("Expect authType: 11 but got: %v", authType)
		}
		authData := make([]byte, msgLen-8)
		_, err = io.ReadFull(c.reader.Reader, authData) // -8: msgLen, authType
		if err != nil {
			log.Printf("Failed to get authData: %v \n", err)
			return err
		}
		_, resp, err = client.Step(authData)
		if err != nil {
			log.Printf("Failed to step 2: %v \n", err)
			return err
		}
		_, err = c.conn.Write(b.buildSASLResponse(resp))
		if err != nil {
			log.Printf("Failed to send SASLResponse: %v \n", err)
			return err
		}
		msgType, err = c.reader.ReadByte()
		if err != nil {
			log.Printf("Failed to read msgType of AuthenticationSASLContinue: %v \n", msgType)
			return err
		}
		if msgType == authMsgType {
			msgLen, err = c.reader.ReadManyBytes(4)
			if err != nil {
				log.Printf("Failed to read msgLen SASLResponse: %v \n", err)
				return err
			}
			authType, err = c.reader.ReadManyBytes(4)
			if err != nil {
				log.Printf("Failed to read authType SASLResponse: %v \n", err)
				return err
			}
			if authType != SASLComplete {
				return fmt.Errorf("Expect authType of SASLResponse is 12 but got: %v", authType)
			}
			authData := make([]byte, msgLen-4)
			_, err = io.ReadFull(c.reader.Reader, authData)
			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("Expect AuthenticationSASLFinal with R letter but got: %v", msgType)
		}
	} else {
		return fmt.Errorf("Expect AuthenticationSASLContinue with R letter but got: %v", msgType)
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
