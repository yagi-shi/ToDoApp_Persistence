package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/todos", todoHandler)
	http.HandleFunc("/todos/edit", editHandler)
	http.HandleFunc("/todos/update", updateHandler)
	http.HandleFunc("/todos/delete", deleteHandler)
	log.Println("server start :8000")
	http.ListenAndServe(":8000", nil)
}
