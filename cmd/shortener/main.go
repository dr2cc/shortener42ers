package main

import (
	"net/http"
	handlers "sh42ers/internal/handler"
	"sh42ers/internal/storage"

	"github.com/go-chi/chi"
)

func main() {
	//mux := http.NewServeMux()
	mux := chi.NewRouter()

	storageInstance := storage.NewURLStorage(make(map[string]string))

	// // Обработчик net/http
	// mux.HandleFunc("POST /{$}", handlers.PostHandler(storageInstance))
	// mux.HandleFunc("GET /{id}", handlers.GetHandler(storageInstance))

	// Обработчик chi
	mux.Post("/", handlers.PostHandler(storageInstance))
	mux.Get("/{id}", handlers.GetHandler(storageInstance))

	http.ListenAndServe(":8080", mux)
}
