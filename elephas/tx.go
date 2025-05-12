package elephas

import (
	"fmt"
)

type Tx struct {
	conn *Connection
}

func NewTransaction(conn *Connection) *Tx {
	return &Tx{conn}
}

func (tx *Tx) Commit() error {
	var b Buffer
	_, err := tx.conn.netConn.Write(b.buildQuery("commit", nil))
	if err != nil {
		return err
	}
	cmdTag, err := tx.conn.reader.ReadCommandComplete()
	if err != nil {
		return err
	}
	if cmdTag != string(commitCmd) {
		return fmt.Errorf("Expect COMMIT command but got (%v)", cmdTag)
	}
	if err := ReadReadyForQuery(tx.conn.reader); err != nil {
		return err
	}
	return nil

}

func (tx *Tx) Rollback() error {
	var b Buffer
	_, err := tx.conn.netConn.Write(b.buildQuery("rollback", nil))
	if err != nil {
		return err
	}
	cmdTag, err := tx.conn.reader.ReadCommandComplete()
	if err != nil {
		return err
	}
	if cmdTag != string(rollbackCmd) {
		return fmt.Errorf("Expect ROLLBACK command but got (%v)", cmdTag)
	}
	if err := ReadReadyForQuery(tx.conn.reader); err != nil {
		return err
	}
	return nil
}
