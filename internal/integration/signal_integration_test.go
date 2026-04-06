package integration

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/agent"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	handlers "github.com/KaziPHone/go-musthave-metrics-tpl/internal/handler"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
)

// TestAgentShutdownAndServerSave интеграционный тест (локальный).
// Проверяет, что при отправке сигнала агент дожидается отправки метрик,
// а сервер сохраняет метрики в файл.
func TestAgentShutdownAndServerSave(t *testing.T) {
	// временный файл для хранилища
	tmpDir := t.TempDir()
	filePath := tmpDir + "/metrics_storage.json"

	// создаём конфигурацию хранилища для сервера
	serverCfg := config.ServerConfig{
		Host:        "",
		FileStorage: filePath,
		// не включаем периодическое автосохранение
		StoreInterval: 0,
		Restore:       false,
		DataBaseDsn:   "",
	}

	st := storage.NewMemStorage(serverCfg)

	// обработчик и роутер
	h := &handlers.Handler{Storage: st, AuditSubject: nil}
	r := chi.NewRouter()
	// нужно распаковка gzip, т.к. агент шлёт сжатые данные
	r.Use(handlers.GzipRequestMiddleware)
	r.Post("/updates/", h.UpdatesHandler)

	// запускаем HTTP-сервер на свободном порту
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	server := &http.Server{Handler: r}
	doneServer := make(chan struct{})
	go func() {
		_ = server.Serve(ln)
		close(doneServer)
	}()

	// регистрируем аккуратную остановку хранилища, чтобы реагировать на сигналы и сохранить файл
	st.StorageGracefulStop(server)

	// готовим конфигурацию агента с адресом сервера
	addr := ln.Addr().String()
	agentCfg := config.AgentConfig{
		Host:           addr,
		ReportInterval: 1,
		PollInterval:   1,
		RateLimit:      1,
	}

	ag := agent.NewAgent(agentCfg)

	stop := make(chan struct{})
	doneAgent := make(chan struct{})
	go func() {
		ag.Start(stop)
		close(doneAgent)
	}()

	// настраиваем обработку сигналов в тесте: при получении сигнала закроем канал stop агента
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigCh
		// close agent stop so it will finish sending
		close(stop)
	}()

	// ждём немного, чтобы агент успел отправить несколько метрик
	time.Sleep(1500 * time.Millisecond)

	// шлём сигнал текущему процессу — хранилище поймает его и сохранит файл,
	// наша тестовая горутина тоже поймает сигнал и закроет канал stop агента
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send signal: %v", err)
	}

	// ждём завершения агента (с таймаутом)
	select {
	case <-doneAgent:
		// ok
	case <-time.After(10 * time.Second):
		t.Fatal("agent did not stop within timeout")
	}

	// ждём завершения сервера (StorageGracefulStop вызовет server.Shutdown)
	select {
	case <-doneServer:
		// ok
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop within timeout")
	}

	// проверяем, что файл хранилища существует и содержит метрики
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read storage file: %v", err)
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		t.Fatalf("failed to unmarshal metrics file: %v", err)
	}

	if len(metrics) == 0 {
		t.Fatalf("expected metrics to be saved, got 0")
	}

	// убираем подписку на сигналы
	signal.Stop(sigCh)
}
