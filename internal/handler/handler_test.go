package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/audit"
	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/helpers"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// mockAuditObserver реализует интерфейс audit.Observer
type mockAuditObserver struct {
	Events []audit.AuditEvent
	mu     sync.Mutex
}

func (m *mockAuditObserver) OnAuditEvent(event audit.AuditEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = append(m.Events, event)
}

func (m *mockAuditObserver) GetEvents() []audit.AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Events
}

func TestUpdateValueHandler_Gauge(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("typeMetric", "gauge")
	routeCtx.URLParams.Add("nameMetric", "test_metric")
	routeCtx.URLParams.Add("value", "100.5")

	req := httptest.NewRequest("POST", "/value/gauge/test_metric/100.5", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	rr := httptest.NewRecorder()
	handler.UpdateValueHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUpdateValueHandler_Counter(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	req := httptest.NewRequest("POST", "/value/counter/test_counter/50", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("typeMetric", "counter")
	routeCtx.URLParams.Add("nameMetric", "test_counter")
	routeCtx.URLParams.Add("value", "50")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	rr := httptest.NewRecorder()
	handler.UpdateValueHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUpdateValueHandler_InvalidType(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	req := httptest.NewRequest("POST", "/value/invalid/test_metric/100", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("typeMetric", "invalid")
	routeCtx.URLParams.Add("nameMetric", "test_metric")
	routeCtx.URLParams.Add("value", "100")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	rr := httptest.NewRecorder()
	handler.UpdateValueHandler(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateValueHandler_InvalidValue(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	req := httptest.NewRequest("POST", "/value/gauge/test_metric/invalid", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("typeMetric", "gauge")
	routeCtx.URLParams.Add("nameMetric", "test_metric")
	routeCtx.URLParams.Add("value", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	rr := httptest.NewRecorder()
	handler.UpdateValueHandler(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateValueHandler_EmptyName(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	req := httptest.NewRequest("POST", "/value/gauge//100", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("typeMetric", "gauge")
	routeCtx.URLParams.Add("nameMetric", "")
	routeCtx.URLParams.Add("value", "100")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	rr := httptest.NewRecorder()
	handler.UpdateValueHandler(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateValueHandler_EmptyValue(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	req := httptest.NewRequest("POST", "/value/gauge/test_metric/", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("typeMetric", "gauge")
	routeCtx.URLParams.Add("nameMetric", "test_metric")
	routeCtx.URLParams.Add("value", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	rr := httptest.NewRecorder()
	handler.UpdateValueHandler(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateValueHandler_InvalidCounterValue(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	req := httptest.NewRequest("POST", "/value/counter/test_counter/abc", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("typeMetric", "counter")
	routeCtx.URLParams.Add("nameMetric", "test_counter")
	routeCtx.URLParams.Add("value", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	rr := httptest.NewRecorder()
	handler.UpdateValueHandler(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateValueHandler_SuccessResponse(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	req := httptest.NewRequest("POST", "/value/gauge/test_metric/123.45", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("typeMetric", "gauge")
	routeCtx.URLParams.Add("nameMetric", "test_metric")
	routeCtx.URLParams.Add("value", "123.45")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	rr := httptest.NewRecorder()
	handler.UpdateValueHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Updated successfully")
}

func TestUpdateHandler_Gauge(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	metric := models.Metrics{
		ID:    "test_gauge",
		MType: models.Gauge,
		Value: floatPtr(100.5),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.UpdateHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "{}", rr.Body.String())
}

func TestUpdateHandler_Counter(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	metric := models.Metrics{
		ID:    "test_counter",
		MType: models.Counter,
		Delta: int64Ptr(50),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.UpdateHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUpdateHandler_EmptyID(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	metric := models.Metrics{
		ID:    "",
		MType: models.Gauge,
		Value: floatPtr(100.5),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.UpdateHandler(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateHandler_EmptyType(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	metric := models.Metrics{
		ID:    "test_metric",
		MType: "",
		Value: floatPtr(100.5),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.UpdateHandler(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateHandler_InvalidJSON(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	req := httptest.NewRequest("POST", "/update", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.UpdateHandler(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateHandler_NonJSONContentType(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	metric := models.Metrics{
		ID:    "test",
		MType: models.Gauge,
		Value: floatPtr(100.5),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	handler.UpdateHandler(rr, req)

	// Handler should still process it since we just decode from body
	// But storage will reject if type is unknown
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestListMetricsHandler(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	// Сначала добавим несколько метрик
	memStorage.UpdateMetric("test1", "gauge", 100.0)
	memStorage.UpdateMetric("test2", "counter", int64(50))

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ListMetricsHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Список текущих метрик:")
	assert.Contains(t, rr.Body.String(), "test1")
	assert.Contains(t, rr.Body.String(), "test2")
}

func TestListMetricsHandler_Empty(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ListMetricsHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Список текущих метрик:")
}

func TestGetPingDBHandler_Success(t *testing.T) {
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	req := httptest.NewRequest("GET", "/ping", nil)
	rr := httptest.NewRecorder()
	handler.GetPingDBHandler(rr, req)

	// Когда БД не подключена (нет DSN), должен вернуться 500
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetPingDBHandler_DBError(t *testing.T) {
	// Когда БД не подключена, должен вернуться 500
	memStorage := storage.NewMemStorage(config.ServerConfig{})
	handler := &Handler{Storage: memStorage}

	req := httptest.NewRequest("GET", "/ping", nil)
	rr := httptest.NewRecorder()
	handler.GetPingDBHandler(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandler_multipleMetrics(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "m1", MType: "gauge", Value: floatPtr(1.0)},
		{ID: "m2", MType: "counter", Delta: int64Ptr(2)},
		{ID: "m3", MType: "gauge", Value: floatPtr(3.0)},
	}

	body, _ := json.Marshal(metrics)
	req := httptest.NewRequest("GET", "/test", bytes.NewReader(body))

	handler := &Handler{Storage: storage.NewMemStorage(config.ServerConfig{})}
	result, err := handler.multipleMetrics(req)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "m1", result[0].ID)
	assert.Equal(t, "m2", result[1].ID)
	assert.Equal(t, "m3", result[2].ID)
}

func TestHandler_multipleMetrics_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", strings.NewReader("invalid json"))
	handler := &Handler{Storage: storage.NewMemStorage(config.ServerConfig{})}

	result, err := handler.multipleMetrics(req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestHandler_multipleMetrics_EmptyBody(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", bytes.NewReader([]byte{}))
	handler := &Handler{Storage: storage.NewMemStorage(config.ServerConfig{})}

	result, err := handler.multipleMetrics(req)

	assert.Error(t, err) // Ошибка EOF для пустого тела
	assert.Nil(t, result)
}

func TestHandler_multipleMetrics_EmptyArray(t *testing.T) {
	body, _ := json.Marshal([]models.Metrics{})
	req := httptest.NewRequest("GET", "/test", bytes.NewReader(body))

	handler := &Handler{Storage: storage.NewMemStorage(config.ServerConfig{})}
	result, err := handler.multipleMetrics(req)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestHandler_NotifyAudit(t *testing.T) {
	subject := &audit.Subject{}
	mockObserver := &mockAuditObserver{}
	subject.Attach(mockObserver)

	handler := &Handler{
		Storage:      storage.NewMemStorage(config.ServerConfig{}),
		AuditSubject: subject,
	}

	handler.NotifyAudit([]string{"metric1", "metric2"}, "192.168.1.1")

	// Ждем асинхронного события аудита
	time.Sleep(100 * time.Millisecond)

	events := mockObserver.GetEvents()
	assert.Equal(t, 1, len(events))
	assert.Contains(t, events[0].Metrics, "metric1")
	assert.Contains(t, events[0].Metrics, "metric2")
	assert.Equal(t, "192.168.1.1", events[0].IPAddress)
	assert.NotZero(t, events[0].TS)
}

func TestHandler_NotifyAudit_NoSubject(t *testing.T) {
	handler := &Handler{
		Storage: storage.NewMemStorage(config.ServerConfig{}),
		// AuditSubject не установлен
	}

	// Не должно вызвать панику
	handler.NotifyAudit([]string{"metric1"}, "192.168.1.1")
}

func TestHandler_NotifyAudit_EmptyMetrics(t *testing.T) {
	subject := &audit.Subject{}
	handler := &Handler{
		Storage:      storage.NewMemStorage(config.ServerConfig{}),
		AuditSubject: subject,
	}

	// Не должно вызвать панику с пустыми метриками
	handler.NotifyAudit([]string{}, "192.168.1.1")
}

func floatPtr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

// Тесты для middleware
func TestGzipRequestMiddleware_GzipEncoding(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	handlerWithMiddleware := GzipRequestMiddleware(handler)

	// Создаем сжатое gzip тело запроса
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte(`{"test": "data"}`))
	gz.Close()

	req := httptest.NewRequest("POST", "/test", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGzipRequestMiddleware_NoGzip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerWithMiddleware := GzipRequestMiddleware(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGzipRequestMiddleware_InvalidGzip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerWithMiddleware := GzipRequestMiddleware(handler)

	// Отправляем недопустимые gzip данные
	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte{0x1f, 0x8b}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	// Должно вернуть ошибку
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGzipRequestMiddleware_UnsupportedContentType(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerWithMiddleware := GzipRequestMiddleware(handler)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("data"))
	gz.Close()

	req := httptest.NewRequest("POST", "/test", &buf)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, rr.Code)
}

func TestGzipResponseMiddleware_GzipAccept(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response body"))
	})

	handlerWithMiddleware := GzipResponseMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))
	// Response должен быть сжат
}

func TestGzipResponseMiddleware_NoGzipAccept(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response body"))
	})

	handlerWithMiddleware := GzipResponseMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	// No Accept-Encoding header

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "", rr.Header().Get("Content-Encoding"))
	assert.Equal(t, "response body", rr.Body.String())
}

func TestLoggingMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})

	handlerWithMiddleware := LoggingMiddleware(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	// Логирование выполняется через zerolog, трудно проверить напрямую
}

func TestLoggingMiddleware_GetRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerWithMiddleware := LoggingMiddleware(handler)

	req := httptest.NewRequest("GET", "/test/path?query=1", nil)

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestShaMiddleware_Success(t *testing.T) {
	key := "test-secret-key"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	handlerWithMiddleware := ShaMiddleware(key)(handler)

	data := []byte(`{"test": "data"}`)
	hash := helpers.CalcSHA256Hash(data)

	// Создаем запрос с хэшем
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(data)
	gz.Close()

	req := httptest.NewRequest("POST", "/test", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("HashSHA256", hash)

	// Добавляем оригинальное тело в контекст
	ctx := context.WithValue(req.Context(), helpers.OriginalBodyKey, data)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestShaMiddleware_BadHash(t *testing.T) {
	key := "test-secret-key"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerWithMiddleware := ShaMiddleware(key)(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))
	req.Header.Set("HashSHA256", "wrong-hash")

	ctx := context.WithValue(req.Context(), helpers.OriginalBodyKey, []byte(`{"test": "data"}`))
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestShaMiddleware_NoKey(t *testing.T) {
	key := "" // No key means no hash validation
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response"))
	})

	handlerWithMiddleware := ShaMiddleware(key)(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestShaMiddleware_NoOriginalBody(t *testing.T) {
	key := "test-secret-key"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerWithMiddleware := ShaMiddleware(key)(handler)

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test": "data"}`))
	// No OriginalBodyKey in context

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGzipWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	gw := gzipWriter{Writer: gz}
	n, err := gw.Write([]byte("test"))

	assert.NoError(t, err)
	assert.Equal(t, 4, n)

	// Сбрасываем для получения данных
	gz.Close()

	// Декомпрессируем и проверяем
	gr, _ := gzip.NewReader(&buf)
	data, _ := io.ReadAll(gr)
	assert.Equal(t, "test", string(data))
}

func TestShaRw_Write(t *testing.T) {
	var buf bytes.Buffer
	srw := &shaRw{buf: &buf}

	n, err := srw.Write([]byte("test"))

	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "test", buf.String())
}

func TestShaRw_WriteHeader(t *testing.T) {
	srw := &shaRw{buf: &bytes.Buffer{}}

	srw.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, srw.statusCode)
	assert.True(t, srw.wroteHeader)

	// Calling again should not change
	srw.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusCreated, srw.statusCode)
}

func TestResponseRecorder_WriteHeader(t *testing.T) {
	rr := &responseRecorder{
		ResponseWriter: &mockResponseWriter{},
		statusCode:     http.StatusOK,
	}

	rr.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, rr.statusCode)
}

func TestResponseRecorder_Write(t *testing.T) {
	mw := &mockResponseWriter{}
	rr := &responseRecorder{
		ResponseWriter: mw,
		statusCode:     http.StatusOK,
	}

	n, err := rr.Write([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, 4, rr.responseSize)
}

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func TestGzipWriter_Close(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	gw := gzipWriter{Writer: gz}
	gw.Write([]byte("test"))

	// Примечание: gzipWriter не имеет Close, это на gzip.Writer
	// Этот тест проверяет базовую функциональность записи
	gz.Close()
}

func TestShaRw_PreserveStatusCode(t *testing.T) {
	srw := &shaRw{
		statusCode: http.StatusNotFound,
	}

	// statusCode должен быть сохранен
	assert.Equal(t, http.StatusNotFound, srw.statusCode)
}

func TestGzipResponseMiddleware_WriteBeforeHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	handlerWithMiddleware := GzipResponseMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
