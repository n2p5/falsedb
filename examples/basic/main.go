package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rjrbt/falsedb"
)

func main() {
	// Open a new database connection
	db, err := falsedb.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, "SELECT * FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		fmt.Println("Found a row!")
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
}
