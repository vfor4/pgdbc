package elephas

import (
	"context"
	"database/sql/driver"
	"net"
)

type Pipeline struct {
	conn net.Conn
	b    Buffer
}

func NewPipeline() *Pipeline {
	// TODO
	return nil
}

func (p Pipeline) QueryContext(ctx context.Context, query string, args []driver.NamedValue) error {
	// _, err := p.conn.Write(p.b.buidParseCmd(query,))
	// return err
	return nil
}

func (p Pipeline) Rows(ctx context.Context) (driver.Rows, error) {
	_, err := p.conn.Write(p.b.buildSync())
	return nil, err
}
