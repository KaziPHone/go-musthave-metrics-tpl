package main

import (
	"log"
	"net/http"

	"github.com/KaziPHone/go-musthave-metrics-tpl/cmd/server/handlers"
	"github.com/KaziPHone/go-musthave-metrics-tpl/cmd/server/storage"
)

func main() {
	storage := storage.NewMemStorage()
	handler := &handlers.Handler{Storage: *storage}
	http.HandleFunc("/update/", handler.UpdateHandler)
	log.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
