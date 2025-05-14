// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://github.com/golang/go/blob/master/src/database/sql/example_test.go
package elephas

import (
	"context"
	"database/sql"
	"log"
	"testing"
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
	db, err = sql.Open("elephas", "postgres://postgres:postgres@localhost:5432/myorder")
	if err != nil {
		log.Fatalf("Failed to connect to user database: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	ec := m.Run()

	if err := db.Close(); err != nil {
		log.Fatalf("Failed to close database: %v", err)
	}
	log.Printf("excode: %v ", ec)
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
		r, err := stmt.ExecContext(context.Background(), v)
		NoError(t, err)
		log.Println(r.RowsAffected())
	}
	// ensureDBValid(t, db)
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
		var n int64
		if err := rows.Scan(&n); err != nil {
			t.Error(err)
		}
	}

	if rows.Err() != nil {
		t.Error(rows.Err())
	}
}

// func TestQueryContext(t *testing.T) {
// 	rows, err := db.QueryContext(ctx, "SELECT name FROM users WHERE age=?", 20)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()
// 	names := make([]string, 0)
//
// 	for rows.Next() {
// 		var name string
// 		if err := rows.Scan(&name); err != nil {
// 			log.Fatal(err)
// 		}
// 		names = append(names, name)
// 	}
// 	rerr := rows.Close()
// 	if rerr != nil {
// 		log.Fatal(rerr)
// 	}
//
// 	if err := rows.Err(); err != nil {
// 		log.Fatal(err)
// 	}
// 	t.Logf("%s are %d years old", strings.Join(names, ", "), 30)
// }

// func TestQueryRowContext(t *testing.T) {
// 	id := 1
// 	var username string
// 	var created time.Time
// 	err := db.QueryRowContext(ctx, "SELECT name, created_at FROM users WHERE user_id=?", id).Scan(&username, &created)
// 	switch {
// 	case err == sql.ErrNoRows:
// 		t.Fatalf("no user with id %d\n", id)
// 	case err != nil:
// 		t.Fatalf("query error: %v\n", err)
// 	default:
// 		t.Logf("username is %q, account created on %s\n", username, created)
// 	}
// }

// func TestExecContext(t *testing.T) {
// 	id := 1
// 	result, err := db.ExecContext(ctx, "UPDATE users SET salary = salary + 3000 WHERE user_id = ?", id)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	rows, err := result.RowsAffected()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	if rows != 1 {
// 		log.Fatalf("expected to affect 1 row, affected %d", rows)
// 	}
// }

// func ExampleDB_Query_multipleResultSets() {
// 	age := 27
// 	q := `
// create temp table uid (id bigint); -- Create temp table for queries.
// insert into uid
// select id from users where age < ?; -- Populate temp table.
//
// -- First result set.
// select
// 	users.id, name
// from
// 	users
// 	join uid on users.id = uid.id
// ;
//
// -- Second result set.
// select
// 	ur.user, ur.role
// from
// 	user_roles as ur
// 	join uid on uid.id = ur.user
// ;
// 	`
// 	rows, err := db.Query(q, age)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()
//
// 	for rows.Next() {
// 		var (
// 			id   int64
// 			name string
// 		)
// 		if err := rows.Scan(&id, &name); err != nil {
// 			log.Fatal(err)
// 		}
// 		log.Printf("id %d name is %s\n", id, name)
// 	}
// 	if !rows.NextResultSet() {
// 		log.Fatalf("expected more result sets: %v", rows.Err())
// 	}
// 	var roleMap = map[int64]string{
// 		1: "user",
// 		2: "admin",
// 		3: "gopher",
// 	}
// 	for rows.Next() {
// 		var (
// 			id   int64
// 			role int64
// 		)
// 		if err := rows.Scan(&id, &role); err != nil {
// 			log.Fatal(err)
// 		}
// 		log.Printf("id %d has role %s\n", id, roleMap[role])
// 	}
// 	if err := rows.Err(); err != nil {
// 		log.Fatal(err)
// 	}
// }

// func TestPingContext(t *testing.T) {
// 	// Ping and PingContext may be used to determine if communication with
// 	// the database server is still possible.
// 	//
// 	// When used in a command line application Ping may be used to establish
// 	// that further queries are possible; that the provided DSN is valid.
// 	//
// 	// When used in long running service Ping may be part of the health
// 	// checking system.
//
// 	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
// 	defer cancel()
//
// 	status := "up"
// 	if err := db.PingContext(ctx); err != nil {
// 		status = "down"
// 	}
// 	log.Println(status)
// }

// func TestPrepare(t *testing.T) {
// 	projects := []struct {
// 		name string
// 		age  int32
// 	}{
// 		{"Person A", 99},
// 		// {"Person B", 24},
// 		// {"Person C", 19},
// 		// {"Person D", 65},test
// 	}
//
// 	stmt, err := db.Prepare("INSERT INTO users(name, age) VALUES(?, ?)")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer stmt.Close() // Prepared statements take up server resources and should be closed after use.
//
// 	for _, project := range projects {
// 		if _, err := stmt.Exec(project.name, project.age); err != nil {
// 			log.Fatal(err)
// 		}
// 	}
// }

// func TestStmtExecContextSuccess(t *testing.T) {
// 	_, err := db.Exec("create temporary table t(id int primary key)")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	stmt, err := db.Prepare("insert into t(id) values ($1::int4)")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer stmt.Close()
//
// 	_, err = stmt.ExecContext(context.Background(), 42)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	// ensureDBValid(t, db)
// }

// func ensureDBValid(t testing.TB, db *sql.DB) {
// 	var sum, rowCount int32
//
// 	rows, err := db.Query("select generate_series(1,$1)", 10)
// 	require.NoError(t, err)
// 	defer rows.Close()
//
// 	for rows.Next() {
// 		var n int32
// 		rows.Scan(&n)
// 		sum += n
// 		rowCount++
// 	}
//
// 	require.NoError(t, rows.Err())
//
// 	if rowCount != 10 {
// 		t.Error("Select called onDataRow wrong number of times")
// 	}
// 	if sum != 55 {
// 		t.Error("Wrong values returned")
// 	}
// }

// func TestStmtExecContextSuccess(t *testing.T) {
//
// 	_, err := db.Exec("create temporary table t(id int primary key)")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	// stmt, err := db.Prepare("insert into t(id) values ($1::int4)")
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }
// 	// defer stmt.Close()
// 	//
// 	// _, err = stmt.ExecContext(context.Background(), 42)
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }
// }

//
// func ExampleTx_Prepare() {
// 	projects := []struct {
// 		mascot  string
// 		release int
// 	}{
// 		{"tux", 1991},
// 		{"duke", 1996},
// 		{"gopher", 2009},
// 		{"moby dock", 2013},
// 	}
//
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer tx.Rollback() // The rollback will be ignored if the tx has been committed later in the function.
//
// 	stmt, err := tx.Prepare("INSERT INTO projects(id, mascot, release, category) VALUES( ?, ?, ?, ? )")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer stmt.Close() // Prepared statements take up server resources and should be closed after use.
//
// 	for id, project := range projects {
// 		if _, err := stmt.Exec(id+1, project.mascot, project.release, "open source"); err != nil {
// 			log.Fatal(err)
// 		}
// 	}
// 	if err := tx.Commit(); err != nil {
// 		log.Fatal(err)
// 	}
// }
//
// func ExampleDB_BeginTx() {
// 	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	id := 37
// 	_, execErr := tx.Exec(`UPDATE users SET status = ? WHERE id = ?`, "paid", id)
// 	if execErr != nil {
// 		_ = tx.Rollback()
// 		log.Fatal(execErr)
// 	}
// 	if err := tx.Commit(); err != nil {
// 		log.Fatal(err)
// 	}
// }
//
// func ExampleConn_ExecContext() {
// 	// A *DB is a pool of connections. Call Conn to reserve a connection for
// 	// exclusive use.
// 	conn, err := db.Conn(ctx)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer conn.Close() // Return the connection to the pool.
// 	id := 41
// 	result, err := conn.ExecContext(ctx, `UPDATE balances SET balance = balance + 10 WHERE user_id = ?;`, id)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	rows, err := result.RowsAffected()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	if rows != 1 {
// 		log.Fatalf("expected single row affected, got %d rows affected", rows)
// 	}
// }
//
// func ExampleTx_ExecContext() {
// 	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	id := 37
// 	_, execErr := tx.ExecContext(ctx, "UPDATE users SET status = ? WHERE id = ?", "paid", id)
// 	if execErr != nil {
// 		if rollbackErr := tx.Rollback(); rollbackErr != nil {
// 			log.Fatalf("update failed: %v, unable to rollback: %v\n", execErr, rollbackErr)
// 		}
// 		log.Fatalf("update failed: %v", execErr)
// 	}
// 	if err := tx.Commit(); err != nil {
// 		log.Fatal(err)
// 	}
// }
//
// func ExampleTx_Rollback() {
// 	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	id := 53
// 	_, err = tx.ExecContext(ctx, "UPDATE drivers SET status = ? WHERE id = ?;", "assigned", id)
// 	if err != nil {
// 		if rollbackErr := tx.Rollback(); rollbackErr != nil {
// 			log.Fatalf("update drivers: unable to rollback: %v", rollbackErr)
// 		}
// 		log.Fatal(err)
// 	}
// 	_, err = tx.ExecContext(ctx, "UPDATE pickups SET driver_id = $1;", id)
// 	if err != nil {
// 		if rollbackErr := tx.Rollback(); rollbackErr != nil {
// 			log.Fatalf("update failed: %v, unable to back: %v", err, rollbackErr)
// 		}
// 		log.Fatal(err)
// 	}
// 	if err := tx.Commit(); err != nil {
// 		log.Fatal(err)
// 	}
// }
//
// func ExampleStmt() {
// 	// In normal use, create one Stmt when your process starts.
// 	stmt, err := db.PrepareContext(ctx, "SELECT username FROM users WHERE id = ?")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer stmt.Close()
//
// 	// Then reuse it each time you need to issue the query.
// 	id := 43
// 	var username string
// 	err = stmt.QueryRowContext(ctx, id).Scan(&username)
// 	switch {
// 	case err == sql.ErrNoRows:
// 		log.Fatalf("no user with id %d", id)
// 	case err != nil:
// 		log.Fatal(err)
// 	default:
// 		log.Printf("username is %s\n", username)
// 	}
// }
//
// func ExampleStmt_QueryRowContext() {
// 	// In normal use, create one Stmt when your process starts.
// 	stmt, err := db.PrepareContext(ctx, "SELECT username FROM users WHERE id = ?")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer stmt.Close()
//
// 	// Then reuse it each time you need to issue the query.
// 	id := 43
// 	var username string
// 	err = stmt.QueryRowContext(ctx, id).Scan(&username)
// 	switch {
// 	case err == sql.ErrNoRows:
// 		log.Fatalf("no user with id %d", id)
// 	case err != nil:
// 		log.Fatal(err)
// 	default:
// 		log.Printf("username is %s\n", username)
// 	}
// }
//
// func ExampleRows() {
// 	age := 27
// 	rows, err := db.QueryContext(ctx, "SELECT name FROM users WHERE age=?", age)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()
//
// 	names := make([]string, 0)
// 	for rows.Next() {
// 		var name string
// 		if err := rows.Scan(&name); err != nil {
// 			log.Fatal(err)
// 		}
// 		names = append(names, name)
// 	}
// 	// Check for errors from iterating over rows.
// 	if err := rows.Err(); err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Printf("%s are %d years old", strings.Join(names, ", "), age)
// }
