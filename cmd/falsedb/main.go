package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]
	port := 8080
	if len(args) >= 2 {
		if args[0] != "-p" {
			fmt.Println("-p is the only valid flag")
			os.Exit(1)
		}
		p, err := strconv.Atoi(args[1])
		if err != nil || 1 >= port || port >= 65535 {
			fmt.Println("port must be a valid integer between 1-65535")
			os.Exit(1)
		}
		port = p
	}

	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Cache-control", "public, max-age=31536000, immutable")
		io.WriteString(w, "")
	}

	http.HandleFunc("/", handlerFunc)
	fmt.Printf("falsedb server is running on port: %v\nyou can run this on an alternate port with the -p or --port flags\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", port), nil))
}
