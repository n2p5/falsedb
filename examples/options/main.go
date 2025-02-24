package main

import (
	"fmt"
	"time"

	"github.com/rjrbt/falsedb"
)

func main() {
	db := falsedb.OpenDB(
		falsedb.WithMaxOpenConns(10),
		falsedb.WithMaxIdleConns(5),
		falsedb.WithConnMaxLifetime(time.Hour),
		falsedb.WithConnMaxIdleTime(30*time.Minute),
	)
	defer db.Close()

	fmt.Println("Database configured with custom connection options, but they are meaningless!")
}
