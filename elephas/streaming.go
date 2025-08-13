package elephas

import (
	"database/sql/driver"
	"errors"
)

func SendDataRowInCopy(c *Connection, args []driver.NamedValue) error {
	var b Buffer
	pn, err := c.reader.Peek(1)
	if err != nil {
		panic(err)
	}
	if pn[0] == copyInResponse {
		_, _ = c.reader.Discard(1)
		_, err := c.reader.Read4Bytes()
		if err != nil {
			panic(err)
		}
		if format, err := c.reader.ReadByte(); err != nil {
			panic(err)
		} else if format == 1 {
			panic("TODO support binary")
		}
		columns, err := c.reader.Read2Bytes()
		if err != nil {
			panic(err)
		}
		_, _ = c.reader.Discard(int(columns) * 2)
		if byten, ok := args[0].Value.([]byte); ok && len(byten) != 0 {
			_, err := c.netConn.Write(b.buildCopyData(byten))
			if err != nil {
				panic(err)
			}
			_, err = c.netConn.Write(b.buildCopyDone())
			if err != nil {
				panic(err)
			}
		} else {
			return errors.New("copy data is nil")
		}
	}
	return nil
}

func ReadDataRowCopyTo(c *Connection)
