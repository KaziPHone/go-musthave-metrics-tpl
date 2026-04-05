package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/audit"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	handlers "github.com/KaziPHone/go-musthave-metrics-tpl/internal/handler"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/helpers"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
)

func main() {
	fmt.Println("Запуск профилирования памяти...")

	// Включаем профилирование памяти
	runtime.GC() // Сборка мусора перед измерением

	// Создаем файл для сохранения профиля
	profileFile, err := os.Create("profiles/base.pprof")
	if err != nil {
		fmt.Printf("Ошибка при создании файла профиля: %v\n", err)
		os.Exit(1)
	}
	defer profileFile.Close()

	// Запускаем профилирование памяти
	if err := pprof.WriteHeapProfile(profileFile); err != nil {
		fmt.Printf("Ошибка при записи профиля кучи: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Профиль кучи сохранен в profiles/base.pprof")

	// Запускаем тесты для нагружения системы
	runLoadTests()

	fmt.Println("Тесты на нагрузку завершены. Запустите: go test -bench=. -memprofile=profiles/result.memprof -benchmem")
}

func runLoadTests() {
	fmt.Println("Запуск тестов на нагрузку...")

	// Тест 1: Storage - множественные обновления
	fmt.Println("Тест 1: Storage.UpdateMetric (1000 итераций)")
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)

	for i := 0; i < 1000; i++ {
		storage.UpdateMetric(fmt.Sprintf("metric_%d", i), "gauge", float64(i))
	}

	for i := 0; i < 1000; i++ {
		storage.UpdateMetric(fmt.Sprintf("counter_%d", i), "counter", int64(i))
	}

	// Тест 2: Handlers - множественные запросы
	fmt.Println("Тест 2: Handler.UpdateValueHandler (500 итераций)")
	h := &handlers.Handler{
		Storage:      storage,
		AuditSubject: nil,
	}

	for i := 0; i < 500; i++ {
		req := httptest.NewRequest("POST", "/update/gauge/test_metric/100.5", nil)
		w := httptest.NewRecorder()
		req.URL.RawPath = fmt.Sprintf("/update/gauge/test_metric_%d/100.5", i)
		req.RequestURI = req.URL.RequestURI()
		h.UpdateValueHandler(w, req)
	}

	// Тест 3: Handlers - пакетное обновление
	fmt.Println("Тест 3: Handler.UpdatesHandler (100 итераций по 10 метрик)")
	metrics := make([]models.Metrics, 10)
	for i := 0; i < 10; i++ {
		metrics[i] = models.Metrics{
			ID:    fmt.Sprintf("test_metric_%d", i),
			MType: "gauge",
			Value: float64Ptr(float64(i)),
		}
	}
	body, _ := json.Marshal(metrics)

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.UpdatesHandler(w, req)
	}

	// Тест 4: SHA256 хэширование (часто вызывается в middleware)
	fmt.Println("Тест 4: SHA256 хэширование (10000 итераций)")
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	for i := 0; i < 10000; i++ {
		_ = helpers.CalcSHA256Hash(data)
	}

	// Тест 5: Gzip сжатие/декомпрессия
	fmt.Println("Тест 5: Gzip сжатие/декомпрессия (500 итераций)")
	for i := 0; i < 500; i++ {
		compressed, _ := compress(data)
		_, _ = decompress(compressed)
	}

	// Тест 6: Audit Notify - просто цикл с созданием события
	fmt.Println("Тест 6: Симуляция уведомления аудита (500 итераций)")
	// Имитация Notify без реальных наблюдателей
	for i := 0; i < 500; i++ {
		_ = audit.AuditEvent{
			TS:        1234567890,
			Metrics:   []string{"metric1", "metric2", "metric3"},
			IPAddress: "127.0.0.1",
		}
	}

	// Тест 7: Цепочка middleware
	fmt.Println("Тест 7: Цепочка middleware (100 итераций)")

	// Подготовка данных для SHA middleware
	originalBody := make([]byte, 1024)
	for i := range originalBody {
		originalBody[i] = byte(i % 256)
	}

	hash := helpers.CalcSHA256Hash(originalBody)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	shaHandler := handlers.ShaMiddleware("test_key")(handler)

	for i := 0; i < 100; i++ {
		body := bytes.NewReader(originalBody)
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("HashSHA256", hash)
		ctx := context.WithValue(req.Context(), helpers.OriginalBodyKey, originalBody)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		shaHandler.ServeHTTP(w, req)
	}

	fmt.Println("Все тесты на нагрузку завершены!")
}

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return nil, fmt.Errorf("ошибка записи в gzip: %v", err)
	}
	if err := gw.Flush(); err != nil {
		return nil, fmt.Errorf("ошибка сброса gzip: %v", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("ошибка закрытия gzip: %v", err)
	}
	return buf.Bytes(), nil
}

func decompress(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	return io.ReadAll(gr)
}

func float64Ptr(v float64) *float64 {
	return &v
}
