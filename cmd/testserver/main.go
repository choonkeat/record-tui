// Command testserver serves files from the current directory over HTTP.
// Port is configured via the PORT environment variable (default 8080).
//
// Usage:
//
//	PORT=3000 go run ./cmd/testserver
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Serving current directory on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, http.FileServer(http.Dir("."))))
}
