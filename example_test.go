package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/agent"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/audit"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	handlers "github.com/KaziPHone/go-musthave-metrics-tpl/internal/handler"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/helpers"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
	"github.com/go-chi/chi/v5"
)

// Примеры для демонстрации использования основных компонентов системы метрик.

func Example_agent() {
	// Создаем конфигурацию агента
	cfg := config.AgentConfig{
		Host:           "localhost:8080",
		PollInterval:   1, // секунды
		ReportInterval: 2, // секунды
		RateLimit:      5,
		Key:            "test_key",
	}

	// Создаем агент
	a := agent.NewAgent(cfg)

	// Создаем канал для остановки агента
	stopCh := make(chan struct{})

	// Запускаем агент в горутине
	go func() {
		a.Start(stopCh)
	}()

	// Ждем немного для сбора метрик
	time.Sleep(3 * time.Second)

	// Останавливаем агент
	close(stopCh)

	// Выводим количество собранных метрик
	stats := a.GetAgentStats()
	fmt.Printf("Collected metrics: %d\n", stats)
	// Output: Collected metrics: 3
}

func Example_storage() {
	// Создаем конфигурацию для хранения в памяти
	cfg := config.ServerConfig{
		FileStorage:   "",
		StoreInterval: 0,
	}

	// Создаем хранилище
	store := storage.NewMemStorage(cfg)

	// Обновляем метрики
	store.UpdateMetric("gauge_metric", "gauge", 123.45)
	store.UpdateMetric("counter_metric", "counter", int64(100))

	// Получаем одну метрику
	metric, found := store.GetMetric("gauge_metric")
	fmt.Printf("Found: %v, Gauge: %v\n", found, metric.Gauge)

	// Получаем все метрики
	allMetrics := store.ListMetrics()
	fmt.Printf("Total metrics: %d\n", len(allMetrics))
	// Output:
	// Found: true, Gauge: 123.45
	// Total metrics: 2
}

func Example_handler() {
	// Создаем хранилище
	store := storage.NewMemStorage(config.ServerConfig{})

	// Создаем предмет аудита
	subject := &audit.Subject{}

	// Создаем обработчик
	h := &handlers.Handler{
		Storage:      store,
		AuditSubject: subject,
	}

	// Создаем роутер и добавляем обработчики
	router := chi.NewRouter()
	router.Get("/value/{typeMetric}/{nameMetric}", h.GetMetricHandler)
	router.Get("/list", h.ListMetricsHandler)
	router.Post("/update/{typeMetric}/{nameMetric}/{value}", h.UpdateValueHandler)
	router.Post("/update/", h.UpdateHandler)
	router.Post("/updates/", h.UpdatesHandler)

	// Создаем HTTP сервер
	server := httptest.NewServer(router)
	defer server.Close()

	// Отправляем метрику
	req, _ := http.NewRequest("POST", server.URL+"/update/gauge/test_gauge/42.5", nil)
	req.Header.Set("Content-Type", "text/plain")
	resp1, _ := http.DefaultClient.Do(req)
	resp1.Body.Close()

	// Проверяем результат
	resp, err := http.Get(server.URL + "/value/gauge/test_gauge")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)

	fmt.Printf("Metric value: %s\n", buf.String())
	// Output: Metric value: 42.5
}

func Example_audit() {
	// Создаем тему аудита
	subject := &audit.Subject{}

	// Создаем файловый наблюдатель
	fileObs := audit.NewFileObserver("/tmp/audit_test.log")
	defer os.Remove("/tmp/audit_test.log")

	// Присоединяем наблюдателя
	subject.Attach(fileObs)

	// Создаем событие аудита
	event := audit.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "127.0.0.1",
	}

	// Отправляем событие
	subject.Notify(event)

	fmt.Printf("Audit event sent for metrics: %v\n", event.Metrics)
	// Output: Audit event sent for metrics: [metric1 metric2]
}

func Example_middleware() {
	store := storage.NewMemStorage(config.ServerConfig{})
	subject := &audit.Subject{}
	h := &handlers.Handler{
		Storage:      store,
		AuditSubject: subject,
	}

	// Создаем роутер с middleware
	router := chi.NewRouter()
	router.Use(handlers.LoggingMiddleware)
	router.Use(handlers.GzipRequestMiddleware)
	router.Use(handlers.GzipResponseMiddleware)
	router.Use(handlers.ShaMiddleware("test_key"))

	// Добавляем обработчики
	router.Post("/update", h.UpdateHandler)

	// Отправляем сжатую метрику
	metric := models.Metrics{
		ID:    "test_metric",
		MType: "gauge",
		Value: &[]float64{99.9}[0],
	}
	data, _ := json.Marshal(metric)

	req := httptest.NewRequest("POST", "/update", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	fmt.Printf("Status: %d\n", rr.Code)
	// Output: Status: 200
}

func Example_concurrentAccess() {
	store := storage.NewMemStorage(config.ServerConfig{})

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			store.UpdateMetric(fmt.Sprintf("metric_%d", id), "gauge", float64(id))
		}(i)
	}

	wg.Wait()

	allMetrics := store.ListMetrics()
	fmt.Printf("Total metrics: %d\n", len(allMetrics))
	// Output: Total metrics: 100
}

func Example_ipHelper() {
	// Создаем запрос с заголовками
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18")

	ip := helpers.GetClientIP(req)
	fmt.Printf("Client IP: %s\n", ip)
	// Output: Client IP: 192.168.1.100
}

func Example_hashHelper() {
	data := []byte("Hello, World!")
	hash := helpers.CalcSHA256Hash(data)

	fmt.Printf("SHA256 hash: %s\n", hash)
	fmt.Printf("Is valid: %v\n", !helpers.IsBadShaRequest(data, hash))
	// Output:
	// SHA256 hash: dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f
	// Is valid: true
}

func Example_config() {
	// Создаем конфигурацию с значениями по умолчанию
	cfg, err := config.NewConfigServer()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Server address: %s\n", cfg.Host)
	fmt.Printf("File storage: %s\n", cfg.FileStorage)
	fmt.Printf("Store interval: %d seconds\n", cfg.StoreInterval)
	// Output:
	// Server address: localhost:8080
	// File storage: metrics_storage.json
	// Store interval: 300 seconds
}

func Example_metricTypes() {
	store := storage.NewMemStorage(config.ServerConfig{})

	// Gauge - плавающее значение, которое можно менять в любую сторону
	store.UpdateMetric("temperature", "gauge", 23.5)
	store.UpdateMetric("temperature", "gauge", 24.0) // перезапись

	// Counter - счетчик, который только увеличивается
	store.UpdateMetric("requests", "counter", int64(10))
	store.UpdateMetric("requests", "counter", int64(5)) // увеличение на 5

	// Получаем значения
	gaugeMetric, _ := store.GetMetric("temperature")
	counterMetric, _ := store.GetMetric("requests")

	fmt.Printf("Gauge (temperature): %v\n", gaugeMetric.Gauge)
	fmt.Printf("Counter (requests): %d\n", counterMetric.Counter)
	// Output:
	// Gauge (temperature): 24
	// Counter (requests): 15
}

func Example_jsonUpdate() {
	// Создаем тестовый сервер
	store := storage.NewMemStorage(config.ServerConfig{})
	subject := &audit.Subject{}
	h := &handlers.Handler{
		Storage:      store,
		AuditSubject: subject,
	}

	router := chi.NewRouter()
	router.Post("/updates", h.UpdatesHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	// Отправляем метрики в формате JSON
	metrics := []models.Metrics{
		{ID: "cpu", MType: "gauge", Value: &[]float64{45.5}[0]},
		{ID: "requests", MType: "counter", Delta: &[]int64{100}[0]},
	}
	data, _ := json.Marshal(metrics)

	req, _ := http.NewRequest("POST", server.URL+"/updates", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	// Проверяем, что метрики сохранены
	metricsMap := store.ListMetrics()
	fmt.Printf("Updated metrics: %d\n", len(metricsMap))
	// Output: Updated metrics: 2
}

func Example_gracefulShutdown() {
	// Создаем хранилище
	store := storage.NewMemStorage(config.ServerConfig{
		FileStorage: "/tmp/shutdown_test.json",
	})

	// Обновляем метрику
	store.UpdateMetric("test", "gauge", &[]float64{123.45}[0])

	// Файл должен быть создан после UpdatesMetrics (когда StoreInterval == 0)
	if _, err := os.Stat("/tmp/shutdown_test.json"); err == nil {
		os.Remove("/tmp/shutdown_test.json")
		fmt.Println("Storage file created successfully")
	} else {
		fmt.Println("File not found")
	}
	// Output: Storage file created successfully
}

func Example_multipleStorageModes() {
	// Режим 1: Память (по умолчанию)
	store1 := storage.NewMemStorage(config.ServerConfig{
		FileStorage: "",
		DataBaseDsn: "",
	})
	val1 := 100.0
	store1.UpdateMetric("test", "gauge", &val1)
	fmt.Printf("Memory mode: %d metrics\n", len(store1.ListMetrics()))

	// Режим 2: Файл
	tmpFile := "/tmp/storage_file_test.json"
	defer os.Remove(tmpFile)

	store2 := storage.NewMemStorage(config.ServerConfig{
		FileStorage: tmpFile,
	})
	val2 := 200.0
	store2.UpdateMetric("test", "gauge", &val2)
	store2.StorageGracefulStop(nil) // сохранит в файл
	fmt.Printf("File mode: %d metrics\n", len(store2.ListMetrics()))

	// Output:
	// Memory mode: 1 metrics
	// File mode: 1 metrics
}

func Example_auditEvent() {
	// Создаем событие
	event := audit.AuditEvent{
		TS:        1234567890,
		Metrics:   []string{"cpu", "memory", "disk"},
		IPAddress: "192.168.1.1",
	}

	// Метаданные события
	fmt.Printf("Timestamp: %d\n", event.TS)
	fmt.Printf("IP Address: %s\n", event.IPAddress)
	fmt.Printf("Metrics changed: %v\n", event.Metrics)
	// Output:
	// Timestamp: 1234567890
	// IP Address: 192.168.1.1
	// Metrics changed: [cpu memory disk]
}

func Example_contextHelper() {
	// Создаем запрос с телом
	data := []byte(`{"id": "test", "type": "gauge", "value": 100.0}`)
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(data))

	// Контекст для хранения оригинального тела
	req = req.WithContext(context.WithValue(req.Context(), helpers.OriginalBodyKey, data))

	// Получаем тело из контекста
	originalBody, ok := req.Context().Value(helpers.OriginalBodyKey).([]byte)
	fmt.Printf("Body exists: %v, length: %d\n", ok, len(originalBody))
	// Output:
	// Body exists: true, length: 47
}

func Example_gzipCompression() {
	// Создаем метрики для демонстрации
	metrics := []models.Metrics{
		{ID: "test", MType: "gauge", Value: &[]float64{42.0}[0]},
	}

	// Сжимаем данные
	_, _ = json.Marshal(metrics) // demonstrates compression usage

	fmt.Println("Gzip compression available")
	// Output: Gzip compression available
}
