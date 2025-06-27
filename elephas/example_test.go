// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go/blob/master/src/database/sql/example_test.go
package elephas

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"testing"
	"time"
)

var (
	ctx context.Context = context.Background()
	db  *sql.DB
)

type order_entity struct {
	id   uint64
	name string
}

func TestMain(m *testing.M) {
	log.SetFlags(log.Lshortfile)

	var err error
	db, err = sql.Open("elephas", "postgres://postgres:postgres@localhost:5432/gosqltest")
	if err != nil {
		log.Fatalf("Failed to connect to user database: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	_ = m.Run()
	if err := db.Close(); err != nil {
		log.Fatalf("Failed to close database: %v", err)
	}
}

func NoError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func xTestStmtExecContextSuccess(t *testing.T) {
	_, err := db.Exec("create temporary table t(id int primary key)")
	NoError(t, err)

	stmt, err := db.Prepare("insert into t(id) values (?)")
	NoError(t, err)
	defer stmt.Close()
	values := []int32{42}
	for _, v := range values {
		_, err := stmt.ExecContext(context.Background(), v)
		NoError(t, err)
	}
}

func xTestStmtQueryContextSucess(t *testing.T) {
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

func xTestConnQuery(t *testing.T) {
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

func xTestQueryNull(t *testing.T) {
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

func xTestTxLifeCycle(t *testing.T) {
	_, err := db.Exec("create temporary table t(c varchar not null)")
	NoError(t, err)
	tx, err := db.BeginTx(context.Background(), nil)
	NoError(t, err)
	value := "a"
	_, err = tx.Exec("insert into t values (?)", value)
	NoError(t, err)
	err = tx.Rollback()
	NoError(t, err)
	var s string
	if err = db.QueryRow("select c from t where c = ?", value).Scan(&s); err != sql.ErrNoRows {
		t.Error("Expected ErrNoRows ")
	}
	_, err = db.Exec("insert into t(c) values (?)", value)
	NoError(t, err)
	err = tx.Commit()
	NoError(t, err)
	err = db.QueryRow("select c from t where c = ?", value).Scan(&s)
	NoError(t, err)
	if s != "a" {
		t.Errorf("Expected %v but got %v", value, s)
	}
}

func xTestQueryInvalidSystax(t *testing.T) {
	// TODO
}

func xTestMultipleStatements(t *testing.T) {
	// conn, err := pgx.Connect(context.Background(), "postgres://postgres:postgres@localhost:5432/gosqltest")
	_, err := db.Exec("create temporary table t(c varchar not null)")
	NoError(t, err)
	_, err = db.Exec(`insert into t values('a'); insert into t values('b')`)
	NoError(t, err)
	var s int32
	err = db.QueryRow("select count(*) from t").Scan(&s)
	NoError(t, err)
	if s != 2 {
		log.Println(s)
		NoError(t, errors.New("Expected 2 records"))
	}
}

func TestImplicitTx(t *testing.T) {
	_, err := db.Exec("create temporary table t(id int unique, c varchar not null)")
	NoError(t, err)
	r, err := db.Exec(`
		INSERT INTO t VALUES(1,'a');
		INSERT INTO t VALUES(2,'b');
		INSERT INTO t VALUES(3,'b');
		`)
	re, err := r.RowsAffected()
	NoError(t, err)
	Equals(t, "RowsEffected", 3, re)
	var s string
	err = db.QueryRow("select c from t").Scan(&s)
	NoError(t, err)
	Equals(t, "TestImplictiTx", "a", s)
}

// func BenchmarkExec(b *testing.B) {
// 	// conn, err := db.Conn(context.Background())
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_, err := db.Query("SELECT 1")
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// }

func xTestIdleConn(t *testing.T) {
	controllerConn, err := sql.Open("elephas", "postgres://postgres:postgres@localhost:5432/gosqltest")
	NoError(t, err)

	db, err := sql.Open("elephas", "postgres://postgres:postgres@localhost:5432/gosqltest")
	NoError(t, err)

	var conns []*sql.Conn
	for range 3 {
		c, err := db.Conn(context.Background())
		NoError(t, err)
		conns = append(conns, c)
	}
	for _, c := range conns {
		err = c.Close()
		NoError(t, err)
	}
	err = controllerConn.PingContext(context.Background())
	NoError(t, err)

	time.Sleep(time.Second)
	if db.Stats().OpenConnections != 2 {
		log.Println(db.Stats().OpenConnections)
		t.Error(errors.New("Expected 1 connections"))
	}
}

func xTestConnWithoutClose(t *testing.T) {
	db, err := sql.Open("elephas", "postgres://postgres:postgres@localhost:5432/gosqltest")
	conn, err := db.Conn(context.Background())
	Equals(t, "open", 1, db.Stats().OpenConnections)
	NoError(t, err)

	err = conn.PingContext(context.Background())
	Equals(t, "inuse", 1, db.Stats().InUse)

	conn2, err := db.Conn(context.Background())
	conn2.PingContext(context.Background())
	Equals(t, "inuse", 2, db.Stats().InUse)
	Equals(t, "open", 2, db.Stats().OpenConnections)
}

func xTestConnWithConnClose(t *testing.T) {
	db, err := sql.Open("elephas", "postgres://postgres:postgres@localhost:5432/gosqltest")
	conn, err := db.Conn(context.Background())
	Equals(t, "open", 1, db.Stats().OpenConnections)
	NoError(t, err)

	err = conn.PingContext(context.Background())
	Equals(t, "inuse", 1, db.Stats().InUse)
	conn.Close()
	Equals(t, "idle after close, return to pool", 1, db.Stats().Idle)

	conn2, err := db.Conn(context.Background())
	conn2.PingContext(context.Background())
	Equals(t, "inuse", 1, db.Stats().InUse)
	Equals(t, "open", 1, db.Stats().OpenConnections)
	Equals(t, "idle", 1, db.Stats().Idle)
}

func xTestCtxDoneWhileWaitingConnToReturnPool(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	db, _ := sql.Open("elephas", "postgres://postgres:postgres@localhost:5432/gosqltest")
	db.SetMaxOpenConns(1)
	_, err := db.Conn(ctx)
	_, err = db.Conn(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected deadline")
	}
}

func xTestWaitingIdleConnAndAbleToGrabIt(t *testing.T) {
	db, _ := sql.Open("elephas", "postgres://postgres:postgres@localhost:5432/gosqltest")
	db.SetMaxOpenConns(1)
	conn, err := db.Conn(ctx)
	NoError(t, err)
	Equals(t, "open", 1, db.Stats().OpenConnections)
	Equals(t, "inuse", 1, db.Stats().InUse)
	Equals(t, "idle", 0, db.Stats().Idle)
	conn.PingContext(context.Background())

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
