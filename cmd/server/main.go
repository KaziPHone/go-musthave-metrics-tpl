package main

import (
	"net"
	"net/http"

	"crypto/rsa"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/audit"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/buildinfo"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	handlers "github.com/KaziPHone/go-musthave-metrics-tpl/internal/handler"
	cryptopkg "github.com/KaziPHone/go-musthave-metrics-tpl/pkg/crypto"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func main() {

	log.Printf("Build version: %s", buildinfo.GetVersion())
	log.Printf("Build date: %s", buildinfo.GetDate())
	log.Printf("Build commit: %s", buildinfo.GetCommit())

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

	// Если указан путь до приватного ключа — загружаем его и ставим middleware для расшифровки
	// перед gzip-мидлваром, чтобы дальше тело было уже расшифровано.
	var privKey *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		k, err := cryptopkg.LoadPrivateKeyFromFile(cfg.CryptoKey)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load crypto private key")
		}
		privKey = k
		router.Use(handlers.DecryptRequestMiddleware(privKey))
	}

	router.Use(handlers.LoggingMiddleware)
	router.Use(handlers.GzipRequestMiddleware)
	router.Use(handlers.GzipResponseMiddleware)
	router.Use(handlers.ShaMiddleware(cfg.Key))

	// Разбираем доверенную подсеть (если указана) и передаём в Handler
	var trustedNet *net.IPNet
	if cfg.TrustedSubnet != "" {
		_, ipnet, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			log.Fatal().Err(err).Msgf("invalid trusted_subnet: %s", cfg.TrustedSubnet)
		}
		trustedNet = ipnet
		log.Info().Msgf("Trusted subnet set to %s", cfg.TrustedSubnet)
	}

	h := &handlers.Handler{
		Storage:      storage.NewMemStorage(*cfg),
		AuditSubject: auditSubject,
		TrustedSubnet: trustedNet,
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

	// Устанавливаем graceful shutdown для хранилища (оно само вызовет server.Shutdown при сигнале)
	h.Storage.StorageGracefulStop(server)

	log.Printf("Starting server on: %s...", cfg.Host)
	// Используем server.ListenAndServe чтобы Shutdown повлиял на этот экземпляр
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Err(err).Msg("server error")
	}
}
