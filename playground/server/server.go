package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("The server is running on port 8080")
	http.ListenAndServe(":8080", http.FileServer(http.Dir("../www")))
}
