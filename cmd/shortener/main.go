package main

import (
	"net/http"
	handlers "sh42ers/internal/handler"
	"sh42ers/internal/storage"
)

func main() {
	mux := http.NewServeMux()
	storageInstance := storage.NewStorage()

	mux.HandleFunc("POST /{$}", handlers.PostHandler(storageInstance))

	mux.HandleFunc("GET /{id}", storageInstance.GetHandler)

	http.ListenAndServe(":8080", mux)
}
