package main

import (
	"fmt"
	"net/http"
	"log"
)

func main() {
	log.Print("Starting [pipeline-runner] service")
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "It's Alive!")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
