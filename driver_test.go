package elephas

import (
	"database/sql"
	"log"
	"testing"
)

func TestMain(m *testing.M) {
	log.SetFlags(log.Lshortfile)

	var err error
	db, err = sql.Open("pgdbc", "postgres://postgres:postgres@localhost:5432/gosqltest")
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

func TestDbOpenInRep(t *testing.T) {
	db, err := sql.Open("pgdbc", "postgresql://postgres:postgres@localhost:5432/pglogrepl?replication=database")
	NoError(t, err)
	NoError(t, db.PingContext(ctx))
}
