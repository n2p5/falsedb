package main

import (
	"context"
	"database/sql"
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

	// Execute a query
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, "SELECT * FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Iterate over results
	for rows.Next() {
		// Process row...
		fmt.Println("Found a row!")
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
}
