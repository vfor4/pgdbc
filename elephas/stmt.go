package elephas

import (
	"bufio"
	"context"
	"database/sql/driver"
	"net"
)

type Stmt struct {
	netConn    net.Conn
	portalName string
}

func (st Stmt) Close() error {
	return nil
}

func (st Stmt) NumInput() int {
	return 0
}

func (st Stmt) Exec(args []driver.Value) (driver.Result, error) {
	// _, err = st.netConn.Write(b.buildExecuteCmd(st.portalName))
	// if err != nil {
	// 	return nil, err
	// }
	panic("use ExecContext")
}

func (st Stmt) Query(args []driver.Value) (driver.Rows, error) {
	panic("use QueryContext")
}

func (st Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	var b Buffer
	// _, err := st.netConn.Write(b.buildFlushCmd())
	// if err != nil {
	// 	return nil, err
	// }
	_, err := st.netConn.Write(b.buildDescribe(st.portalName))
	if err != nil {
		return nil, err
	}
	_, err = st.netConn.Write(b.buildSync())
	if err != nil {
		return nil, err
	}
	reader := NewReader(bufio.NewReader(st.netConn))
	rows, err := ReadSimpleQueryRes(reader)
	if err != nil {
		return nil, err
	}
	rows.conn = st.netConn

	return &Rows{}, nil
}
