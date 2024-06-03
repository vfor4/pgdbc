package order

import (
	"database/sql"
	"fmt"
	"net/http"
)

func RegisterHandlers() {
	handler := new(orderHandler)
	http.Handle("GET /order/{id}", handler)
}

type orderHandler struct{}

func (oh orderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sql.NullString
}
