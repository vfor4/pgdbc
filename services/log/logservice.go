package log

import (
	"io"
	stdlog "log"
	"net/http"
	"os"
	"time"
)

var log *stdlog.Logger

type logPath string

func (lp logPath) Write(p []byte) (n int, err error) {
	f, err := os.OpenFile(time.Now().Format("1999-12-01 12:22:22"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, nil
	}
	defer f.Close()
	return f.Write(p)
}

func Run(path string) {
	log = stdlog.New(logPath(path), "", stdlog.LstdFlags)
}

func RegisterHandler() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		msg, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("%v\n", msg)
	})
}
