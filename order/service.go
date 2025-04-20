package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "order/elephas"
	"time"
)

func RegisterHandlers() {
	http.Handle("GET /order/{id}", http.HandlerFunc(getOrderHandler))
	http.Handle("POST /order", http.HandlerFunc(postOrderHandler))
}

func getOrderHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	order := getOrder(id)
	w.Write([]byte(order.String()))
}

func postOrderHandler(w http.ResponseWriter, r *http.Request) {
	var o order
	err := json.NewDecoder(r.Body).Decode(&o)
	defer r.Body.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	addOrder(o)
}

type order struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (o order) String() string {
	return fmt.Sprintf("id: %v; name: %v", o.Id, o.Name)
}

var db *sql.DB

func init() {
	var dsn = "postgres://postgres:postgres@localhost:5432/myorder"
	var err error
	db, err = sql.Open("elephas", dsn)
	if err != nil {
		panic(err)
	}
}

func getOrder(id string) order {
	ctx, cancle := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancle()
	var o order
	db.Ping()
	err := db.QueryRowContext(ctx, "select id, name from order_table_z where id = ?", id).Scan(&o.Id, &o.Name)
	if err != nil {
		log.Println(err)
		return order{}
	}
	return o
}

func addOrder(o order) (int64, error) {
	ctx, cancle := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancle()
	rs, err := db.ExecContext(ctx, "insert into order_table values (?, ?) returning id", o.Id, o.Name)
	if err != nil {
		return 0, err
	}
	return rs.LastInsertId()
}
