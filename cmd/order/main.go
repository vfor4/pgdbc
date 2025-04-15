package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	_ "order/elephas"
	"order/order"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type Order struct {
	ID   int
	Name string
}

func main() {
	ctx, err := order.Start(context.Background(), "localhost", "8081")
	if err != nil {
		fmt.Println(err)
	}
	<-ctx.Done()
}

func toPrepare(db *sql.DB, ctx context.Context) error {
	printErr := func(err error) error {
		fmt.Println(err)
		return err
	}
	stmt, err := db.PrepareContext(ctx, "select * from orders where id = $1")
	if err != nil {
		return printErr(err)
	}
	rows, err := stmt.QueryContext(context.Background())
	if err != nil {
		return printErr(err)
	}
	fmt.Println(rows)

	return nil
}

func toWrite(db *sql.DB, ctx context.Context) (int64, error) {
	// Create a helper function for preparing failure results.
	fail := func(err error) (int64, error) {
		return 0, fmt.Errorf("CreateOrder: %v", err)
	}
	// Get a Tx for making transaction requests.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}

	// Confirm that album inventory is enough for the order.
	var enough bool
	const (
		productName = "iphone"
		quantity    = 5
	)
	if err = tx.QueryRowContext(ctx, "SELECT (quantity >= ?) as enough from orders where name = ?", quantity, productName).
		Scan(&enough); err != nil {
		if err == sql.ErrNoRows {
			return fail(fmt.Errorf("no such product"))
		}
		return fail(err)
	}
	if !enough {
		return fail(fmt.Errorf("not enough inventory"))
	} else {
		log.Printf("we have more than %v %v(s) in stock\n", quantity, productName)
	}
	_, err = tx.ExecContext(ctx, "UPDATE orders SET quantity = quantity - ? WHERE name = ?", quantity, productName)
	if err != nil {
		tx.Rollback()
		return fail(err)
	} else {
		tx.Commit()
	}
	//
	// // Create a new row in the album_order table.
	// result, err := tx.ExecContext(ctx, "INSERT INTO orders (id, name, quantity, date) VALUES (?, ?, ?, ?)",
	// 	15, "samsung", 20, time.Now())
	// if err != nil {
	// 	return fail(err)
	// }
	// // Get the ID of the order item just created.
	// orderID, err := result.LastInsertId()
	// if err != nil {
	// 	return fail(err)
	// }
	//
	// // Commit the transaction.
	// if err = tx.Commit(); err != nil {
	// 	return fail(err)
	// }
	//
	// // Return the order ID.
	// return orderID, nil
	//
	return 1, nil
}
