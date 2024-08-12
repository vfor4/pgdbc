package main

import (
	"context"
	"database/sql"
	"log"
	_ "order/elephas"
	"time"
)

type Order struct {
	ID   int
	Name string
}

func main() {
	var (
		orders []*Order
		dsn    = "postgres://postgres:postgres@localhost:5432/record"
	)
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	db, err := sql.Open("elephas", dsn)
	if err != nil {
		log.Fatal(err)
	}
	status := "up"
	if err := db.PingContext(ctx); err != nil {
		status = "ops"
	}
	log.Println(status)
	rows, err := db.Query("SELECT * FROM orders")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var o Order

		if err := rows.Scan(&o.ID, &o.Name); err != nil {
			log.Fatal(err)
		}

		orders = append(orders, &o)
	}

	if err := rows.Err(); err != nil {
		log.Fatal("main ", err)
	}
	log.Println(orders[0], orders[1])
}
