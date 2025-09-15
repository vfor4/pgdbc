// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go/blob/master/src/database/sql/example_test.go
package elephas

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"log"
	"runtime/debug"
	"testing"
	"time"
)

var (
	ctx, _ = context.WithTimeout(context.Background(), time.Duration(5*time.Minute))
	db     *sql.DB
)

type order_entity struct {
	id   uint64
	name string
}

func NoError(t *testing.T, err error) {
	if err != nil {
		debug.PrintStack()
		t.Fatal(err)
	}
}

func CloseRealConn(dc any) error {
	if c, ok := dc.(driver.Conn); ok {
		return c.Close()
	}
	return nil
}

func YesError(t *testing.T, err error) {
	if err == nil {
		debug.PrintStack()
		t.Fatal("Expected error")
	}
}

func TestStmtExecContextSuccess(t *testing.T) {
	conn, _ := db.Conn(ctx)
	_, err := conn.ExecContext(ctx, "create temporary table t(id int primary key)")
	NoError(t, err)

	stmt, err := conn.PrepareContext(ctx, "insert into t(id) values (?)")
	NoError(t, err)
	defer stmt.Close()
	values := []int32{42}
	for _, v := range values {
		_, err := stmt.ExecContext(context.Background(), v)
		NoError(t, err)
	}
}

func TestStmtQueryContextSucess(t *testing.T) {
	stmt, err := db.Prepare("select * from generate_series(1,?) n")
	NoError(t, err)
	defer stmt.Close()

	var end int32
	end = 5
	rows, err := stmt.QueryContext(context.Background(), end)
	NoError(t, err)

	for rows.Next() {
		var n int32
		if err := rows.Scan(&n); err != nil {
			t.Error(err)
		}
	}

	if rows.Err() != nil {
		t.Error(rows.Err())
	}
}

func TestConnQuery(t *testing.T) {
	rows, err := db.Query("select 'foo', n from generate_series(?, ?) n", int32(1), int32(10))
	NoError(t, err)
	rc := int32(0)
	for rows.Next() {
		rc++
		var f string
		var n int32
		err := rows.Scan(&f, &n)
		NoError(t, err)
		if f != "foo" {
			t.Errorf("Expected 'foo', got %v", f)
		}
		if n != rc {
			t.Errorf("Expected %d but got %d", n, rc)
		}
	}
	NoError(t, rows.Err())

	err = rows.Close()
	NoError(t, err)
}

func TestQueryNull(t *testing.T) {
	rows, err := db.Query("select ?", nil)
	NoError(t, err)

	for rows.Next() {
		var n sql.NullInt64
		err := rows.Scan(&n)
		NoError(t, err)
		if n.Valid != false {
			t.Errorf("Expected Null but got %v", n)
		}
	}
}

func TestTxLifeCycle(t *testing.T) {
	conn, _ := db.Conn(ctx)
	_, err := conn.ExecContext(ctx, "create temporary table t(c varchar not null)")
	NoError(t, err)
	tx, err := conn.BeginTx(context.Background(), nil)
	NoError(t, err)
	value := "a"
	_, err = tx.Exec("insert into t values (?)", value)
	NoError(t, err)
	err = tx.Rollback()
	NoError(t, err)
	var s string
	if err = conn.QueryRowContext(ctx, "select c from t where c = ?", value).Scan(&s); err != sql.ErrNoRows {
		t.Error("Expected ErrNoRows")
	}
	_, err = conn.ExecContext(ctx, "insert into t(c) values (?)", value)
	NoError(t, err)
	err = conn.QueryRowContext(ctx, "select c from t where c = ?", value).Scan(&s)
	NoError(t, err)
	if s != "a" {
		t.Errorf("Expected %v but got %v", value, s)
	}
}

func TestOpenTxTwice(t *testing.T) {
	_, _ = db.Exec("truncate test")

	tx, err := db.BeginTx(context.Background(), nil)
	NoError(t, err)
	_, err = tx.Exec("insert into test values(1,'a')")
	NoError(t, err)
	err = tx.Commit()
	NoError(t, err)

	tx2, err := db.BeginTx(context.Background(), nil)
	NoError(t, err)
	_, err = tx2.Exec("insert into test values(2, 'b')")
	NoError(t, err)
	err = tx2.Commit()
	NoError(t, err)

	var i int
	err = db.QueryRow("select count(id) from test").Scan(&i)
	NoError(t, err)
	Equals(t, "open tx twice", 2, i)
}

func TestMultipleStatements(t *testing.T) {
	conn, _ := db.Conn(ctx)
	_, err := conn.ExecContext(ctx, "create temporary table t(c varchar not null)")
	NoError(t, err)
	_, err = conn.ExecContext(ctx, `insert into t values('a'); insert into t values('b')`)
	NoError(t, err)
	var s int32
	err = conn.QueryRowContext(ctx, "select count(*) from t").Scan(&s)
	NoError(t, err)
	if s != 2 {
		log.Println(s)
		NoError(t, errors.New("Expected 2 records"))
	}
}

func TestMultpleExtendedQuery(t *testing.T) {
	conn, _ := db.Conn(ctx)
	_, err := conn.ExecContext(ctx, "create temporary table t(id int unique, c varchar not null)")
	NoError(t, err)
	r, err := conn.ExecContext(ctx, `
		INSERT INTO t VALUES(1,'a');
		INSERT INTO t VALUES(2,'b');
		INSERT INTO t VALUES(3,'b');
		`)
	re, err := r.RowsAffected()
	NoError(t, err)
	Equals(t, "RowsEffected", 3, re)
	var s string
	err = conn.QueryRowContext(ctx, "select c from t").Scan(&s)
	NoError(t, err)
	Equals(t, "TestImplictiTx", "a", s)
}

func TestImplitTx(t *testing.T) {
	conn, _ := db.Conn(ctx)
	_, err := conn.ExecContext(ctx, "create temporary table t(id int unique, c varchar not null)")
	NoError(t, err)
	_, err = conn.ExecContext(ctx, `
		INSERT INTO t VALUES(1,'a');
		INSERT INTO t VALUES(1,'b');
		INSERT INTO t VALUES(3,'b');
		`)
	YesError(t, err)
}

func TestExplitcitTx(t *testing.T) {
	_, err := db.Exec("truncate test")
	NoError(t, err)
	_, err = db.Exec(`
		BEGIN;
		INSERT INTO test VALUES(1, 'a');
		COMMIT;
		INSERT INTO test VALUES(2, 'b');
		SELECT 1/0;
		`)
	YesError(t, err)
	var s string
	err = db.QueryRow("select n from test where id = ?", 1).Scan(&s)
	Equals(t, "TestImplictiTx", "a", s)
}

func TestIdleConn(t *testing.T) {
	controllerConn, err := sql.Open("pgdbc", "postgres://postgres:postgres@localhost:5432/gosqltest")
	NoError(t, err)

	db, err := sql.Open("pgdbc", "postgres://postgres:postgres@localhost:5432/gosqltest")
	NoError(t, err)

	var conns []*sql.Conn
	for range 3 {
		c, err := db.Conn(ctx)
		NoError(t, err)
		conns = append(conns, c)
	}

	for _, c := range conns {
		err = c.Close()
		NoError(t, err)
	}
	err = controllerConn.PingContext(ctx)
	NoError(t, err)
	Equals(t, "Expected 2 connections", db.Stats().OpenConnections, 2)
}

func TestConnWithoutClose(t *testing.T) {
	db, err := sql.Open("pgdbc", "postgres://postgres:postgres@localhost:5432/gosqltest")
	conn, err := db.Conn(ctx)
	Equals(t, "open", 1, db.Stats().OpenConnections)
	NoError(t, err)

	err = conn.PingContext(ctx)
	Equals(t, "inuse", 1, db.Stats().InUse)

	conn2, err := db.Conn(ctx)
	conn2.PingContext(ctx)
	Equals(t, "inuse", 2, db.Stats().InUse)
	Equals(t, "open", 2, db.Stats().OpenConnections)
}

func TestConnWithConnClose(t *testing.T) {
	db, err := sql.Open("pgdbc", "postgres://postgres:postgres@localhost:5432/gosqltest")
	conn, err := db.Conn(ctx)
	Equals(t, "open", 1, db.Stats().OpenConnections)
	NoError(t, err)

	err = conn.PingContext(ctx)
	Equals(t, "inuse", 1, db.Stats().InUse)
	conn.Close()
	Equals(t, "closing call, no more in use", 0, db.Stats().InUse)
	Equals(t, "closing call, return to pool", 1, db.Stats().Idle)
}

func TestCtxDoneWhileWaitingConnToReturnPool(t *testing.T) {
	ctx, ccFn := context.WithTimeout(context.Background(), time.Second)
	defer ccFn()
	db, _ := sql.Open("pgdbc", "postgres://postgres:postgres@localhost:5432/gosqltest")
	db.SetMaxOpenConns(1)
	_, err := db.Conn(ctx)
	_, err = db.Conn(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected deadline")
	}
}

func TestWaitingIdleConnAndAbleToGrabIt(t *testing.T) {
	db, _ := sql.Open("pgdbc", "postgres://postgres:postgres@localhost:5432/gosqltest")
	db.SetMaxOpenConns(1)
	conn, err := db.Conn(ctx)
	conn.PingContext(context.Background())
	NoError(t, err)
	Equals(t, "open", 1, db.Stats().OpenConnections)
	Equals(t, "inuse", 1, db.Stats().InUse)
	Equals(t, "idle", 0, db.Stats().Idle)

	done := make(chan struct{})
	go func() {
		log.Println("Waiting...")
		_, err = db.Conn(context.Background())
		NoError(t, err)
		done <- struct{}{}
	}()
	time.Sleep(2 * time.Second)
	conn.Close()
	<-done
}

func Equals[V comparable](t *testing.T, msg string, expect, actual V) {
	if expect != actual {
		t.Fatalf("(%s)- Expected %v, got %v\n", msg, expect, actual)
	}
}
