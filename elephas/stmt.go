package elephas

import (
	"context"
	"database/sql/driver"
	"net"
)

type Stmt struct {
	netConn   net.Conn
	namedStmt string
}

func (st Stmt) Close() error {
	return nil
}

func (st Stmt) NumInput() int {
	return 0
}

func (st Stmt) Exec(args []driver.Value) (driver.Result, error) {
	panic("use ExecContext")
}

func (st Stmt) Query(args []driver.Value) (driver.Rows, error) {
	panic("use QueryContext")
}

func (st Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	var b Buffer
	st.netConn.Write(b.buidBindCmd(st.namedStmt))
	st.netConn.Write(b.buildExecuteCmd("testportal"))
	return &Rows{}, nil
}
