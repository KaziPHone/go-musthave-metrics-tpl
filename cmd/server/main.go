package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	handlers "github.com/KaziPHone/go-musthave-metrics-tpl/internal/handler"
)

func main() {
	cfg, err := config.NewConfigServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	router := chi.NewRouter()

	router.Use(handlers.LoggingMiddleware)
	router.Use(handlers.GzipRequestMiddleware)
	router.Use(handlers.GzipResponseMiddleware)

	h := &handlers.Handler{Storage: storage.NewMemStorage(*cfg)}
	defer h.Storage.Close()

	configString, _ := json.Marshal(*cfg)
	log.Info().Str("config", string(configString)).Msg("Config loaded")

	router.Get("/", h.ListMetricsHandler)
	router.Get("/value/{typeMetric}/{nameMetric}", h.GetMetricHandler)
	router.Get("/ping", h.GetPingDBHandler)

	router.Post("/update/{typeMetric}/{nameMetric}/{value}", h.UpdateValueHandler)
	router.Post("/update/", h.UpdateHandler)
	router.Post("/updates/", h.UpdatesHandler)
	router.Post("/value/", h.ValueMetricHandler)

	server := &http.Server{
		Addr:    cfg.Host,
		Handler: router,
	}

	// Start server
	go func() {
		log.Info().Msgf("Starting server on: %s...", cfg.Host)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("Server start failed")
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)
	<-stopChan

	// Shutdown server
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Server shutdown failed")
	}
	log.Info().Msg("Server shutdown successfully")
}
