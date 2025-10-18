package main

import (
	"log"
	"net/http"

	"github.com/KaziPHone/go-musthave-metrics-tpl/cmd/server/handlers"
	"github.com/KaziPHone/go-musthave-metrics-tpl/cmd/server/storage"
	"github.com/go-chi/chi/v5"
)

func main() {

	router := chi.NewRouter()
	storage := storage.NewMemStorage()
	handler := &handlers.Handler{Storage: *storage}

	router.Get("/", handler.ListMetricsHandler)
	router.Get("/value/{typeMetric}/{nameMetric}", handler.GetMetricHandler)
	router.Post("/update/{typeMetric}/{nameMetric}/{value}", handler.UpdateHandler)

	log.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(err)
	}
}
