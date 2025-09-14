package elephas

import (
	"os"
	"testing"
)

func TestCopyOneRowNoError(t *testing.T) {
	conn, _ := db.Conn(ctx)
	defer conn.Raw(CloseRealConn)
	_, err := conn.ExecContext(ctx, "create temporary table t(id int not null, n varchar not null)")
	NoError(t, err)
	byten := []byte("1\tAlice")
	_, err = conn.ExecContext(ctx, "COPY t(id,n) from STDIN", byten)
	NoError(t, err)
	var id int
	var s string
	err = conn.QueryRowContext(ctx, "select * from t where id = ?", 1).Scan(&id, &s)
	Equals(t, "copy testing id", 1, id)
	Equals(t, "copy testing n", "Alice", s)
}

func TestCopyFromCSV(t *testing.T) {
	conn, _ := db.Conn(ctx)
	defer conn.Raw(CloseRealConn)
	_, err := conn.ExecContext(ctx, "create temporary table t(id int not null, n varchar not null)")
	NoError(t, err)
	f, err := os.ReadFile("copy.csv")
	NoError(t, err)
	_, err = conn.ExecContext(ctx, "COPY t(id,n) from STDIN (format csv)", f)
	NoError(t, err)
	var id int
	var s string
	err = conn.QueryRowContext(ctx, "select * from t where id = ?", 6).Scan(&id, &s)
	Equals(t, "copy testing id", 6, id)
	Equals(t, "copy testing n", "Frank", s)
}
