package elephas

import (
	"fmt"
	"io"
	"log"
)

type Tx struct {
	conn *Connection
}

func NewTransaction(conn *Connection) *Tx {
	return &Tx{conn}
}

func (tx *Tx) Commit() error {
	if err := CheckReadyForQuery(tx.conn.reader, InTx); err != nil {
		return err
	}
	var b Buffer
	_, err := tx.conn.netConn.Write(b.buildQuery("commit", nil))
	if err != nil {
		return err
	}
	t, err := tx.conn.reader.ReadByte()
	if err != nil {
		return err
	}
	if t == commandComplete {
		tags, err := ReadCommandComplete(tx.conn.reader)
		if err != nil && err != io.EOF {
			return err
		}
		if tags[0] != string(commitCmd) {
			return fmt.Errorf("Expect COMMIT command but got (%v)", tags)
		}
	}
	return nil

}

func (tx *Tx) Rollback() error {
	if err := CheckReadyForQuery(tx.conn.reader, Idle); err != nil {
		return err
	}
	var b Buffer
	_, err := tx.conn.netConn.Write(b.buildQuery("rollback", nil))
	if err != nil {
		return err
	}
	t, err := tx.conn.reader.ReadByte()
	if err != nil {
		return err
	}
	log.Println("test type: ", t)
	if t == commandComplete {
		tags, err := ReadCommandComplete(tx.conn.reader)
		if err != nil && err != io.EOF {
			return err
		}
		if tags[0] != string(rollbackCmd) {
			return fmt.Errorf("Expect ROLLBACK command but got (%v)", tags)
		}
	}
	return nil
}
