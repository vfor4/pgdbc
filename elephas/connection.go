package elephas

import (
	"bufio"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"mellium.im/sasl"
)

// Conn
type Connection struct {
	cfg     *Config
	netConn net.Conn
	reader  *Reader
}

func (c *Connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	var b Buffer
	_, err := c.netConn.Write(b.buildQuery(query, args))
	if err != nil {
		return nil, err
	}
	return Result{reader: c.reader}, nil
}

// PrepareContext returns a prepared statement, bound to this connection.
// context is for the preparation of the statement,
// it must not store the context within the statement itself.
func (c *Connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return c.Prepare(query)
}

func (c *Connection) Prepare(query string) (driver.Stmt, error) {
	var b Buffer
	name := "test"
	cmd := b.buidParseCmd(query, name)
	_, err := c.netConn.Write(cmd)
	if err != nil {
		return nil, err
	}
	portalName := "portalvu"
	cmd = b.buildBindCmd(name, portalName)
	_, err = c.netConn.Write(cmd)
	if err != nil {
		return nil, err
	}
	// cmd = b.buildFlushCmd()
	// _, err = c.netConn.Write(cmd)
	// if err != nil {
	// 	return nil, err
	// }

	return &Stmt{netConn: c.netConn, portalName: portalName}, nil
}

func (c *Connection) Close() error {
	return c.netConn.Close()
}

// deprecated function, use BeginTx instead
func (c *Connection) Begin() (driver.Tx, error) {
	panic("not implemented")
}

func (c *Connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if sql.IsolationLevel(opts.Isolation) != sql.LevelDefault {
		return nil, errors.New("Not implemented")
	}
	if opts.ReadOnly {
		return nil, errors.New("Not implemented")
	}
	var b Buffer
	_, err := c.netConn.Write(b.buildQuery("Begin", nil))
	if err != nil {
		return nil, err
	}
	cmdTag, err := c.reader.ReadCommandComplete()
	if err != nil {
		return nil, fmt.Errorf("Unable to ReadAndExpect(%v)", commandComlete)
	}
	if cmdTag != string(beginCmd) {
		return nil, fmt.Errorf("Expect BEGIN command tag but got (%v)", cmdTag)
	}
	txStatus, err := c.reader.ReadReadyForQuery()
	if err != nil {
		return nil, fmt.Errorf("Expecte to ReadAndExpect(%v)", txStatus)
	}
	if txStatus == T {
		log.Println("in tx")
	}
	return NewTransaction(c), nil
}

// https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-START-UP
func (c *Connection) makeHandShake() error {
	var b Buffer
	if _, err := c.netConn.Write(b.buildStartUpMsg()); err != nil {
		log.Fatalf("Failed to make hande shake: %v", err)
		return err
	}
	for {
		msgType, err := c.reader.ReadByte()
		if err != nil {
			return err
		}
		msgLen, err := c.reader.ReadBytesToUint32()
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
		case backendKeyData:
			c.reader.Discard(int(msgLen - 4))
		case readyForQuery:
			c.reader.Discard(int(msgLen - 4))
			return nil
		default:
			log.Println(string(msgType))
		}
	}
}

func (c *Connection) doAuthentication(b Buffer) error {
	authType, err := c.reader.ReadBytesToUint32()
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
		return err
	}
	_, err = c.netConn.Write(b.buildSASLInitialResponse(resp))
	if err != nil {
		return err
	}
	data, err := c.reader.handleAuthResp(SASLContinue)
	if err != nil {
		return err
	}
	_, resp, err = client.Step(data)
	if err != nil {
		return err
	}

	_, err = c.netConn.Write(b.buildSASLResponse(resp))
	if err != nil {
		return err
	}

	data, err = c.reader.handleAuthResp(SASLComplete)
	if err != nil {
		return err
	}
	if _, _, err = client.Step(data); err != nil {
		return err
	}
	if client.State() != sasl.ValidServerResponse {
		log.Printf("invalid server reponse: %v", client.State())
	}
	_, err = c.reader.handleAuthResp(AuthSuccess)
	if err != nil {
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
	reader := NewReader(bufio.NewReader(dConn))
	conn := Connection{netConn: dConn, reader: reader, cfg: cfg}
	if err := conn.makeHandShake(); err != nil {
		log.Fatalf("Failed to make handle shake: %v", err)
		return nil, err
	}

	return &conn, nil
}

func (c *Connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	var b Buffer
	_, err := c.netConn.Write(b.buildQuery(query, args))
	if err != nil {
		log.Printf("Failed to send Query: %v", err)
		return nil, err
	}
	rows, err := ReadRowDescription(c.reader)
	if err != nil {
		return nil, err
	}
	return (&rows), nil
}
