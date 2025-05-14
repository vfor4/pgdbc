package elephas

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"net"
	"strconv"
)

type Stmt struct {
	netConn   net.Conn
	reader    *Reader
	statement string
	want      int
}

func (st Stmt) Close() error {
	return nil
}

func (st Stmt) NumInput() int {
	return st.want
}

// Deprecated: Drivers should implement StmtExecContext instead (or additionally).
func (st Stmt) Exec(args []driver.Value) (driver.Result, error) {
	panic("use ExecContext")
}

func (st Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if len(args) != 0 && len(args[0].Name) != 0 {
		panic("not suppported named arg")
	}
	portal := "portal_test"
	var b Buffer
	_, err := st.netConn.Write(b.buildBindCmd(args, st.statement, portal))
	if err != nil {
		return nil, err
	}
	_, err = st.netConn.Write(b.buildExecuteCmd(portal))
	if err != nil {
		return nil, err
	}
	_, err = st.netConn.Write(b.buildFlushCmd())
	if err != nil {
		return nil, err
	}
	if bc, err := st.reader.ReadByte(); err != nil {
		return nil, err
	} else if bc != bindComplete {
		return nil, fmt.Errorf("Expected BindComplete but got %v", bc)
	}
	st.reader.Discard(4)
	tags, err := ReadCommandComplete(st.reader)
	if err != nil && err != io.EOF {
		return nil, err
	}
	_, err = st.netConn.Write(b.buildSync())
	if err != nil {
		return nil, err
	}

	if ra, err := strconv.Atoi(tags[len(tags)-1]); err != nil {
		return driver.RowsAffected(0), nil
	} else {
		return driver.RowsAffected(ra), nil
	}
}

func (st Stmt) CheckNamedValue(n *driver.NamedValue) error {
	return nil
}

// Deprecated: Drivers should implement StmtQueryContext instead (or additionally).
func (st Stmt) Query(args []driver.Value) (driver.Rows, error) {
	panic("use QueryContext")
}

func (st Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	var b Buffer
	pn := "query_test"
	_, err := st.netConn.Write(b.buildBindCmd(args, st.statement, pn))
	if err != nil {
		return nil, err
	}
	_, err = st.netConn.Write(b.buildExecuteCmd(pn))
	if err != nil {
		return nil, err
	}
	_, err = st.netConn.Write(b.buildFlushCmd())
	if err != nil {
		return nil, err
	}
	return &Rows{}, nil
}

// func ReadParameterDescription(r *Reader)
