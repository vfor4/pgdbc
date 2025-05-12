package elephas

import (
	"bufio"
	"context"
	"database/sql/driver"
	"net"
)

type Stmt struct {
	netConn   net.Conn
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
	// _, err = st.netConn.Write(b.buildFlushCmd())
	// if err != nil {
	// 	return nil, err
	// }

	_, err = st.netConn.Write(b.buildExecuteCmd(portal))
	if err != nil {
		return nil, err
	}
	// _, err = st.netConn.Write(b.buildFlushCmd())
	// if err != nil {
	// 	return nil, err
	// }
	// _, err = st.netConn.Write(b.buildSync())
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
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
	_, err := st.netConn.Write(b.buildDescribe(st.statement))
	if err != nil {
		return nil, err
	}
	_, err = st.netConn.Write(b.buildSync())
	if err != nil {
		return nil, err
	}
	reader := NewReader(bufio.NewReader(st.netConn))
	_, err = ReadRows(reader, nil)
	if err != nil {
		return nil, err
	}

	return &Rows{}, nil
}
