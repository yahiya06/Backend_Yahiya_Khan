package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT") // R2: use PORT if set
	if port == "" {
		port = "8080"
	}

	srv := NewServer()
	log.Println("listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, srv.Routes()))
}