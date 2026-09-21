package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	filepathRoot := "."
	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))

	port := "8080"
	server := &http.Server{
		Handler: mux,
		Addr: ":" + port,
	}
	
	log.Printf("Starting server on port %s, serving files from %s", port, filepathRoot)
	log.Fatal(server.ListenAndServe())
}