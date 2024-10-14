package order

import (
	"net/http"
)

func RegisterHandlers() {
	handler := new(orderHandler)
	http.Handle("GET /order/{id}", handler)
}

type orderHandler struct{}

func (oh orderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
