package main

import (
	"context"
	"database/sql"
	"fmt"
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
		dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
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
	fmt.Print(status)
	// rows, err := db.Query("SELECT * FROM orders WHERE id=$1 ", id)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for rows.Next() {
	// 	var o Order

	// 	if err := rows.Scan(&o.ID, &o.Name); err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	orders = append(orders, &o)
	// }

	// if err := rows.Err(); err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Print(orders)
}
