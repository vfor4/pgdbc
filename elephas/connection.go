package elephas

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"mellium.im/sasl"
)

// Conn
type Connection struct {
	cfg     *Config
	netConn net.Conn
	reader  *Reader
}

func (c *Connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return c.Prepare(query)
}

var hw hash.Hash = sha256.New()

func (c *Connection) Prepare(query string) (driver.Stmt, error) {
	if err := CheckReadyForQuery(c.reader, Idle); err != nil {
		return nil, err
	}
	var b Buffer
	_, _ = hw.Write([]byte(query))
	name := fmt.Sprintf("%x", hw.Sum(nil))
	w := strings.Count(query, "?")
	for i := range w {
		query = strings.Replace(query, "?", fmt.Sprintf("$%d", i+1), 1)
	}
	_, err := c.netConn.Write(b.buidParseCmd(query, name, w))
	if err != nil {
		return nil, err
	}
	_, err = c.netConn.Write(b.buildFlushCmd())
	if err != nil {
		return nil, err
	}
	err = ReadStmtComplete(c.reader, parseComplete)
	if err != nil {
		return nil, err
	}

	return &Stmt{netConn: c.netConn, reader: c.reader, statement: name, want: w}, nil
}

func (c *Connection) Close() error {
	return c.netConn.Close()
}

// deprecated function, use BeginTx instead
func (c *Connection) Begin() (driver.Tx, error) {
	panic("not implemented")
}

func (c *Connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if err := CheckReadyForQuery(c.reader, Idle); err != nil {
		return nil, err
	}
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
	t, err := c.reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if t == commandComplete {
		cmdTag, err := ReadCommandComplete(c.reader)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if cmdTag[0] != string(beginCmd) {
			return nil, fmt.Errorf("Expect BEGIN command tag but got (%v)", cmdTag)
		}
	}
	return NewTransaction(c), nil
}

// https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-START-UP
func (c *Connection) makeHandShake() error {
	var b Buffer
	if _, err := c.netConn.Write(b.buildStartUpMsg(c.cfg.User, c.cfg.Database)); err != nil {
		log.Fatalf("Failed to make hande shake: %v", err)
		return err
	}
	for {
		msgType, err := c.reader.ReadByte()
		if err != nil {
			return err
		}
		msgLen, err := c.reader.Read4Bytes()
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
		case errorResponseMsg:
			// https://www.postgresql.org/docs/current/protocol-message-formats.html#PROTOCOL-MESSAGE-FORMATS-ERRORRESPONSE
			errResponse := ReadErrorResponse(c.reader)
			panic(fmt.Sprintf("Server response with an error = %+v\n", errResponse.Error()))
		default:
			panic(fmt.Sprintf("Not expected type %v", msgType))
		}
	}
}

func (c *Connection) doAuthentication(b Buffer) error {
	authType, err := c.reader.Read4Bytes()
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
		switch string(data[:len(data)-1]) {
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

	conn := Connection{netConn: dConn, reader: NewReader(bufio.NewReader(dConn)), cfg: cfg}
	if err := conn.makeHandShake(); err != nil {
		log.Fatalf("Failed to make handle shake: %v", err)
		return nil, err
	}

	return &conn, nil
}

func (c *Connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := CheckReadyForQuery(c.reader, Idle); err != nil {
		return nil, err
	}
	var b Buffer
	if _, err := c.netConn.Write(b.buildQuery(query, args)); err != nil {
		return nil, err
	}
	rows, err := ReadRows(c.reader)
	if err != nil {
		return &Row{}, err
	}
	return &rows, nil
}

func (c *Connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if err := CheckReadyForQuery(c.reader, Idle); err != nil {
		return nil, err
	}
	var b Buffer
	_, err := c.netConn.Write(b.buildQuery(query, args))
	if err != nil {
		return nil, err
	}
	r, err := ReadResult(c.reader, query)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Connection) Ping(ctx context.Context) error {
	var b Buffer
	_, err := c.netConn.Write(b.buildQuery("select 1", nil))
	if err != nil {
		return err
	}
	for {
		t, err := c.reader.ReadByte()
		if err != nil {
			return err
		}
		l, err := c.reader.Read4Bytes()
		if err != nil {
			return err
		}
		switch t {
		case rowDescription:
			_, _ = c.reader.Discard(int(l - 4))
		case dataRow:
			if fc, err := c.reader.Read2Bytes(); err != nil {
				return err
			} else if fc != 1 {
				return errors.New("field count is not 1")
			}
			if cl, err := c.reader.Read4Bytes(); err != nil {
				return err
			} else if cl != 1 {
				return errors.New("column length is not 1")
			}
			if d, err := c.reader.ReadByte(); err != nil {
				return err
			} else if d != 49 {
				return fmt.Errorf("Expected 1 but got %v", string(d))
			}
		case commandComplete:
			_, _ = c.reader.Discard(int(l - 4))
		case readyForQuery:
			_, _ = c.reader.Discard(int(l - 4))
			return nil
		default:
			z, _ := c.reader.ReadByte()
			panic(z)
		}
	}
}
