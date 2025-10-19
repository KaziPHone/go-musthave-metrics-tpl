package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/handlers"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
	"github.com/go-chi/chi/v5"
)

func main() {

	serverHost := flag.String("a", "localhost:8080", "адрес HTTP-сервера")
	flag.Parse()

	router := chi.NewRouter()
	memStorage := storage.NewMemStorage()
	h := &handlers.Handler{Storage: memStorage}

	router.Get("/", h.ListMetricsHandler)
	router.Get("/value/{typeMetric}/{nameMetric}", h.GetMetricHandler)
	router.Post("/update/{typeMetric}/{nameMetric}/{value}", h.UpdateHandler)

	log.Printf("Starting server on:%s...", *serverHost)
	err := http.ListenAndServe(*serverHost, router)
	if err != nil {
		log.Fatal(err)
	}
}
