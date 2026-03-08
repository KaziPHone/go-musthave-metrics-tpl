package main

import (
	"net/http"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/audit"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	handlers "github.com/KaziPHone/go-musthave-metrics-tpl/internal/handler"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func main() {

	cfg, err := config.NewConfigServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Инициализация системы аудита
	var auditSubject *audit.Subject

	// Если включён хотя бы один приёмник — создаём субъект
	if cfg.AuditFile != "" || cfg.AuditURL != "" {
		auditSubject = &audit.Subject{}
	}

	// Подключаем файловый наблюдатель
	if cfg.AuditFile != "" {
		fileObs := audit.NewFileObserver(cfg.AuditFile)
		auditSubject.Attach(fileObs)
		log.Info().Msgf("Audit to file enabled: %s", cfg.AuditFile)
	}

	// Подключаем HTTP наблюдатель
	if cfg.AuditURL != "" {
		httpObs := audit.NewHTTPObserver(cfg.AuditURL)
		auditSubject.Attach(httpObs)
		log.Info().Msgf("Audit to URL enabled: %s", cfg.AuditURL)
	}

	router := chi.NewRouter()

	router.Use(handlers.LoggingMiddleware)
	router.Use(handlers.GzipRequestMiddleware)
	router.Use(handlers.GzipResponseMiddleware)
	router.Use(handlers.ShaMiddleware(cfg.Key))

	h := &handlers.Handler{
		Storage:      storage.NewMemStorage(*cfg),
		AuditSubject: auditSubject,
	}

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

	h.Storage.StorageGracefulStop(server)

	log.Printf("Starting server on: %s...", cfg.Host)
	err = http.ListenAndServe(cfg.Host, router)
	if err != nil {
		log.Err(err)
	}
}
