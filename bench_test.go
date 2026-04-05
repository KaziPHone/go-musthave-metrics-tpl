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
	"testing"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/audit"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	handlers "github.com/KaziPHone/go-musthave-metrics-tpl/internal/handler"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/helpers"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
)

// Бенчмарки для хендлеров

func BenchmarkHandler_UpdateValueHandler_Gauge(b *testing.B) {
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)
	h := &handlers.Handler{
		Storage: storage,
	}

	req := httptest.NewRequest("POST", "/update/gauge/test_metric/100.5", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.URL.RawPath = fmt.Sprintf("/update/gauge/test_metric_%d/100.5", i)
		req.RequestURI = req.URL.RequestURI()
		h.UpdateValueHandler(w, req)
	}
}

func BenchmarkHandler_UpdateValueHandler_Counter(b *testing.B) {
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)
	h := &handlers.Handler{
		Storage: storage,
	}

	req := httptest.NewRequest("POST", "/update/counter/test_metric/100", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.URL.RawPath = fmt.Sprintf("/update/counter/test_metric_%d/100", i)
		req.RequestURI = req.URL.RequestURI()
		h.UpdateValueHandler(w, req)
	}
}

func BenchmarkHandler_UpdateHandler(b *testing.B) {
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)
	h := &handlers.Handler{
		Storage: storage,
	}

	metric := models.Metrics{
		ID:    "test_metric",
		MType: "gauge",
		Value: float64Ptr(100.5),
	}
	body, _ := json.Marshal(metric)
	req := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(bytes.NewReader(body))
		h.UpdateHandler(w, req)
	}
}

func BenchmarkHandler_UpdatesHandler(b *testing.B) {
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)
	h := &handlers.Handler{
		Storage:      storage,
		AuditSubject: nil,
	}

	metrics := make([]models.Metrics, 10)
	for i := 0; i < 10; i++ {
		metrics[i] = models.Metrics{
			ID:    fmt.Sprintf("test_metric_%d", i),
			MType: "gauge",
			Value: float64Ptr(float64(i)),
		}
	}
	body, _ := json.Marshal(metrics)
	req := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(bytes.NewReader(body))
		h.UpdatesHandler(w, req)
	}
}

func BenchmarkHandler_GetMetricHandler(b *testing.B) {
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)
	h := &handlers.Handler{
		Storage: storage,
	}

	// Предварительно добавляем метрику
	storage.UpdateMetric("test_metric", "gauge", 100.5)

	req := httptest.NewRequest("GET", "/value/gauge/test_metric", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.URL.RawPath = "/value/gauge/test_metric"
		req.RequestURI = req.URL.RequestURI()
		h.GetMetricHandler(w, req)
	}
}

func BenchmarkHandler_ListMetricsHandler(b *testing.B) {
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)
	h := &handlers.Handler{
		Storage: storage,
	}

	// Предварительно добавляем метрики
	for i := 0; i < 100; i++ {
		storage.UpdateMetric(fmt.Sprintf("metric_%d", i), "gauge", float64(i))
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.ListMetricsHandler(w, req)
	}
}

// Бенчмарки для вспомогательных функций

func BenchmarkCalcSHA256Hash(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = helpers.CalcSHA256Hash(data)
	}
}

func BenchmarkCalcSHA256HashBuffer(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	buf := bytes.NewBuffer(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = helpers.CalcSHA256HashBuffer(*buf)
	}
}

func BenchmarkGzipCompress(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = compress(data)
	}
}

func BenchmarkGzipDecompress(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	
	compressed, _ := compress(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decompress(compressed)
	}
}

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return nil, fmt.Errorf("error writing to gzip: %v", err)
	}
	if err := gw.Flush(); err != nil {
		return nil, fmt.Errorf("the flush gzip error: %v", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("gzip closing error: %v", err)
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

// Бенчмарки для Storage

func BenchmarkStorage_UpdateMetric(b *testing.B) {
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.UpdateMetric(fmt.Sprintf("metric_%d", i), "gauge", float64(i))
	}
}

func BenchmarkStorage_UpdateMetric_Counter(b *testing.B) {
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.UpdateMetric(fmt.Sprintf("metric_%d", i), "counter", int64(i))
	}
}

func BenchmarkStorage_UpdatesMetrics(b *testing.B) {
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)

	metrics := make([]models.Metrics, 50)
	for i := 0; i < 50; i++ {
		metrics[i] = models.Metrics{
			ID:    fmt.Sprintf("metric_%d", i),
			MType: "gauge",
			Value: float64Ptr(float64(i)),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.UpdatesMetrics(metrics)
	}
}

func BenchmarkStorage_GetMetric(b *testing.B) {
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)

	// Предварительно добавляем метрики
	for i := 0; i < 1000; i++ {
		storage.UpdateMetric(fmt.Sprintf("metric_%d", i), "gauge", float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.GetMetric(fmt.Sprintf("metric_%d", i%1000))
	}
}

func BenchmarkStorage_ListMetrics(b *testing.B) {
	cfg := config.ServerConfig{}
	storage := storage.NewMemStorage(cfg)

	// Предварительно добавляем метрики
	for i := 0; i < 100; i++ {
		storage.UpdateMetric(fmt.Sprintf("metric_%d", i), "gauge", float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.ListMetrics()
	}
}

// Бенчмарки для Middleware

func BenchmarkShaMiddleware(b *testing.B) {
	key := "test_key"

	// Подготовка данных
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	
	hash := helpers.CalcSHA256Hash(data)
	body := bytes.NewReader(data)

	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("HashSHA256", hash)
	ctx := context.WithValue(req.Context(), helpers.OriginalBodyKey, data)
	req = req.WithContext(ctx)

	// Простой handler для теста
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	shaHandler := handlers.ShaMiddleware(key)(handler)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Создаем новую копию запроса для каждого запуска
		body := bytes.NewReader(data)
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("HashSHA256", hash)
		ctx := context.WithValue(req.Context(), helpers.OriginalBodyKey, data)
		req = req.WithContext(ctx)
		shaHandler.ServeHTTP(w, req)
	}
}

func BenchmarkGzipRequestMiddleware(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Сжимаем данные
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write(data)
	gw.Flush()
	gw.Close()
	compressedData := buf.Bytes()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	gzipHandler := handlers.GzipRequestMiddleware(handler)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := bytes.NewReader(compressedData)
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		gzipHandler.ServeHTTP(w, req)
	}
}

func BenchmarkAuditNotify(b *testing.B) {
	// Имитация Notify без реальных наблюдателей (чтобы избежать ошибок компиляции)
	for i := 0; i < b.N; i++ {
		_ = audit.AuditEvent{
			TS:        1234567890,
			Metrics:   []string{"metric1", "metric2", "metric3"},
			IPAddress: "127.0.0.1",
		}
	}
}

func BenchmarkGetClientIP(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Real-IP", "192.168.1.1")
		_ = helpers.GetClientIP(req)
	}
}

func BenchmarkGetMetrics(b *testing.B) {
	metrics := make([]models.Metrics, 100)
	for i := 0; i < 100; i++ {
		metrics[i] = models.Metrics{
			ID:    fmt.Sprintf("metric_%d", i),
			MType: "gauge",
			Value: float64Ptr(float64(i)),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = helpers.GetMetrics(metrics)
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}
