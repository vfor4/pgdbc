package elephas

import (
	"os"
	"testing"
)

func TestCopyOneRowNoError(t *testing.T) {
	_, err := db.Exec("create temporary table t(id int not null, n varchar not null)")
	NoError(t, err)
	byten := []byte("1\tAlice")
	_, err = db.Exec("COPY t(id,n) from STDIN", byten)
	NoError(t, err)
	var id int
	var s string
	err = db.QueryRow("select * from t where id = ?", 1).Scan(&id, &s)
	Equals(t, "copy testing id", 1, id)
	Equals(t, "copy testing n", "Alice", s)
}

func TestCopyFromCSV(t *testing.T) {
	_, err := db.Exec("create temporary table t(id int not null, n varchar not null)")
	NoError(t, err)
	f, err := os.ReadFile("copy.csv")
	NoError(t, err)
	_, err = db.Exec("COPY t(id,n) from STDIN (format csv)", f)
	NoError(t, err)
	var id int
	var s string
	err = db.QueryRow("select * from t where id = ?", 6).Scan(&id, &s)
	Equals(t, "copy testing id", 6, id)
	Equals(t, "copy testing n", "Frank", s)
}

func TestCopyToOneRowNoError(t *testing.T) {
	_, err := db.Exec("create temporary table t(id int not null, n varchar not null); insert into t(id, n) values(1, 'zalo')")
	NoError(t, err)
	_, err = db.Exec("copy t(id,n) to STDOUT")
	NoError(t, err)
}
