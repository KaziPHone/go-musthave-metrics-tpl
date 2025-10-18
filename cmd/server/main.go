package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/KaziPHone/go-musthave-metrics-tpl/cmd/server/handlers"
	"github.com/KaziPHone/go-musthave-metrics-tpl/cmd/server/storage"
	"github.com/go-chi/chi/v5"
)

func main() {

	serverHost := flag.String("a", "localhost:8080", "адрес HTTP-сервера")
	flag.Parse()

	router := chi.NewRouter()
	storage := storage.NewMemStorage()
	handler := &handlers.Handler{Storage: *storage}

	router.Get("/", handler.ListMetricsHandler)
	router.Get("/value/{typeMetric}/{nameMetric}", handler.GetMetricHandler)
	router.Post("/update/{typeMetric}/{nameMetric}/{value}", handler.UpdateHandler)

	log.Printf("Starting server on:%s...", *serverHost)
	err := http.ListenAndServe(*serverHost, router)
	if err != nil {
		log.Fatal(err)
	}
}
