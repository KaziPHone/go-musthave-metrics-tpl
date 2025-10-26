package main

import (
	"log"
	"net/http"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/handlers"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
	"github.com/go-chi/chi/v5"
)

func main() {

	cfg := config.NewConfigServer()

	router := chi.NewRouter()
	memStorage := storage.NewMemStorage()
	h := &handlers.Handler{Storage: memStorage}

	router.Get("/", h.ListMetricsHandler)
	router.Get("/value/{typeMetric}/{nameMetric}", h.GetMetricHandler)
	router.Post("/update/{typeMetric}/{nameMetric}/{value}", h.UpdateHandler)

	log.Printf("Starting server on:%s...", cfg.Host)
	err := http.ListenAndServe(cfg.Host, router)
	if err != nil {
		log.Fatal(err)
	}
}
