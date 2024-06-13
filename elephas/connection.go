package elephas

import (
	"bufio"
	"context"
	"database/sql/driver"
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
	return c.conn.Close()
}

func (c *Connection) Begin() (driver.Tx, error) {
	return nil, nil
}

// https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-START-UP
func (c *Connection) makeHandShake() error {
	var b Buffer
	if _, err := c.conn.Write(b.buildStartUpMsg()); err != nil {
		log.Fatalf("Failed to make hande shake: %v", err)
		return err
	}
	log.Println("Sent StartupMessage")
	for {
		msgType, err := c.reader.ReadByte()
		if err != nil {
			return err
		}
		msgLen, err := c.reader.ReadManyBytes(4)
		if err != nil {
			return err
		}
		switch msgType {
		case authMsgType:
			if err = c.doAuthentication(b); err != nil {
				log.Printf("Failed to do authentication: %v", err)
				return err
			}
		case parameterStatus:
			// https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-ASYNC
			c.reader.Discard(int(msgLen - 4))
			log.Println("parameterStatus")
		case backendKeyData:
			c.reader.Discard(int(msgLen - 4))
			log.Println("backendKeyData")
		case readyForQuery:
			log.Println("readyForQuery")
		default:
			log.Println(string(msgType))
		}
	}
}

func (c *Connection) doAuthentication(b Buffer) error {
	authType, err := c.reader.ReadManyBytes(4)
	if err != nil {
		return err
	}
	if authType == SASL {
		data, err := c.reader.ReadBytes(0)
		c.reader.ReadByte() // get  rid of last byte
		if err != nil {
			log.Fatalf("Failed to handle AuthenticationSASL; %v", err)
			return err
		}
		switch string(string(data[:len(data)-1])) {
		case sasl.ScramSha256.Name:
			log.Println("Start SASL authentication")
			err = c.authSASL(&b)
			if err != nil {
				return err
			}
			log.Println("Finish SASL authentication")
		default:
			panic("TODO ScramSha256Plus")
		}
	} else {
		panic("TODO SASL")
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
	data, err := c.reader.handleAuthResp(SASLContinue)
	if err != nil {
		log.Println("Failed to handle AuthenticationSASLContinue")
		return err
	}
	_, resp, err = client.Step(data)
	if err != nil {
		log.Printf("Failed to step: %v \n", err)
		return err
	}

	_, err = c.conn.Write(b.buildSASLResponse(resp))
	if err != nil {
		log.Printf("Failed to send SASLResponse (Step4): %v \n", err)
		return err
	}

	data, err = c.reader.handleAuthResp(SASLComplete)
	if err != nil {
		log.Printf("Failed to handle AuthenticationSASLFinal (complete): %v", err)
		return err
	}
	if _, _, err = client.Step(data); err != nil {
		log.Printf("client.Step 3 failed: %v", err)
		return err
	}
	if client.State() != sasl.ValidServerResponse {
		log.Printf("invalid server reponse: %v", client.State())
	}
	_, err = c.reader.handleAuthResp(AuthSuccess)
	if err != nil {
		log.Printf("Failed to handle AuthenticationSASLFinal (success): %v", err)
		return err
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
	if err := conn.makeHandShake(); err != nil {
		log.Fatalf("Failed to make handle shake: %v", err)
		return nil, err
	}

	return &conn, nil
}
