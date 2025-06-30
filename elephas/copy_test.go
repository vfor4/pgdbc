package elephas

import (
	"testing"
)

func TestCopyNoError(t *testing.T) {
	byten := []byte("1a")
	_, err := db.Exec("COPY test(id, n) from STDIN", byten)
	NoError(t, err)
}
