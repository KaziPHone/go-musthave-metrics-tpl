package main

import (
	"net/http"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	handlers "github.com/KaziPHone/go-musthave-metrics-tpl/internal/handler"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func main() {

	cfg := config.NewConfigServer()

	router := chi.NewRouter()

	router.Use(handlers.LoggingMiddleware)
	router.Use(handlers.GzipRequestMiddleware)
	router.Use(handlers.GzipResponseMiddleware)

	h := &handlers.Handler{Storage: storage.NewMemStorage(*cfg)}

	router.Get("/", h.ListMetricsHandler)
	router.Get("/value/{typeMetric}/{nameMetric}", h.GetMetricHandler)
	router.Get("/ping", h.GetPingDBHandler)

	router.Post("/update/{typeMetric}/{nameMetric}/{value}", h.UpdateValueHandler)
	router.Post("/update/", h.UpdateHandler)
	router.Post("/value/", h.ValueMetricHandler)

	server := &http.Server{
		Addr: cfg.Host,
		Handler: router,
	}

	h.Storage.GracefulStop(server)

	log.Printf("Starting server on: %s...", cfg.Host)
	err := http.ListenAndServe(cfg.Host, router)
	if err != nil {
		log.Err(err)
	}
}
