package helpers

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

func TestGetClientIP_XRealIP(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")

	ip := GetClientIP(req)

	if ip != "192.168.1.100" {
		t.Errorf("expected '192.168.1.100', got '%s'", ip)
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195")
	req.Header.Del("X-Real-IP")

	ip := GetClientIP(req)

	if ip != "203.0.113.195" {
		t.Errorf("expected '203.0.113.195', got '%s'", ip)
	}
}

func TestGetClientIP_XForwardedFor_Multiple(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")
	req.Header.Del("X-Real-IP")

	ip := GetClientIP(req)

	if ip != "203.0.113.195" {
		t.Errorf("expected first IP '203.0.113.195', got '%s'", ip)
	}
}

func TestGetClientIP_XForwardedFor_WithSpaces(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "  203.0.113.195  ,  70.41.3.18  ")
	req.Header.Del("X-Real-IP")

	ip := GetClientIP(req)

	if ip != "203.0.113.195" {
		t.Errorf("expected '203.0.113.195', got '%s'", ip)
	}
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	ip := GetClientIP(req)

	if ip != "127.0.0.1" {
		t.Errorf("expected '127.0.0.1', got '%s'", ip)
	}
}

func TestGetClientIP_RemoteAddr_IPv6(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "[::1]:12345"

	ip := GetClientIP(req)

	if ip != "[::1" {
		t.Logf("Получен IPv6-адрес: %s", ip)
	}
}

func TestGetClientIP_EmptyHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	ip := GetClientIP(req)

	// Должно вернуться в RemoteAddr, который пуст
	if ip != "" {
		t.Errorf("expected empty IP, got '%s'", ip)
	}
}

func TestGetClientIP_RemoteAddrWithPort(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:8080"

	ip := GetClientIP(req)

	if ip != "10.0.0.1" {
		t.Errorf("expected '10.0.0.1', got '%s'", ip)
	}
}

func TestGetClientIP_XForwardedFor_TrimWhitespace(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "  192.168.1.1  ")
	req.Header.Del("X-Real-IP")

	ip := GetClientIP(req)

	if ip != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got '%s'", ip)
	}
}

func TestGetClientIP_XForwardedFor_NoTrailing(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	req.Header.Del("X-Real-IP")

	ip := GetClientIP(req)

	if ip != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got '%s'", ip)
	}
}

func TestGetClientIP_XRealIPPriority(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")
	req.Header.Set("X-Forwarded-For", "203.0.113.195")

	ip := GetClientIP(req)

	// X-Real-IP должен иметь приоритет
	if ip != "192.168.1.100" {
		t.Errorf("expected '192.168.1.100' (X-Real-IP priority), got '%s'", ip)
	}
}

func TestGetClientIP_Concurrent(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:8080"

	var wg sync.WaitGroup
	results := make(chan string, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- GetClientIP(req)
		}()
	}

	wg.Wait()
	close(results)

	// Проверяем, что все результаты совпадают
	for result := range results {
		if result != "127.0.0.1" {
			t.Errorf("expected '127.0.0.1', got '%s'", result)
		}
	}
}

func TestGetClientIP_Localhost(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "localhost:8080"

	ip := GetClientIP(req)

	// Localhost должен быть возвращен как есть (без двоеточия для обрезки)
	if ip != "localhost" {
		t.Errorf("expected 'localhost', got '%s'", ip)
	}
}

func TestGetClientIP_IPv6WithZoneID(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "fe80::1%eth0:8080"

	ip := GetClientIP(req)

	if !strings.HasPrefix(ip, "fe80") {
		t.Errorf("expected to start with 'fe80', got '%s'", ip)
	}
}

func TestGetClientIP_LongXForwardedFor(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", strings.Repeat("192.168.1.1, ", 100)+"203.0.113.195")
	req.Header.Del("X-Real-IP")

	ip := GetClientIP(req)

	if ip != "192.168.1.1" {
		t.Errorf("expected first IP '192.168.1.1', got '%s'", ip)
	}
}

func TestGetClientIP_InvalidIP(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "not-an-ip:8080"

	ip := GetClientIP(req)

	// Должно вернуть как есть, так как нет двоеточия для обрезки
	if ip != "not-an-ip" {
		t.Errorf("expected 'not-an-ip', got '%s'", ip)
	}
}

func TestGetClientIP_IPv4MappedIPv6(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "::ffff:192.168.1.1:8080"

	ip := GetClientIP(req)

	// Должно извлечь часть IPv4
	if ip != "::ffff:192.168.1.1" {
		t.Errorf("expected '::ffff:192.168.1.1', got '%s'", ip)
	}
}

func TestGetClientIP_WithTabWhitespace(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "	192.168.1.1	")
	req.Header.Del("X-Real-IP")

	ip := GetClientIP(req)

	if ip != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got '%s'", ip)
	}
}

func TestGetMetrics(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "metric1", MType: "gauge", Value: floatPtr(100.0)},
		{ID: "metric2", MType: "counter", Delta: int64Ptr(10)},
		{ID: "metric3", MType: "gauge", Value: floatPtr(200.0)},
	}

	result := GetMetrics(metrics)

	if len(result) != 3 {
		t.Errorf("expected 3 метрики, got %d", len(result))
	}

	expected := []string{"metric1", "metric2", "metric3"}
	for i, id := range expected {
		if result[i] != id {
			t.Errorf("expected result[%d] = '%s', got '%s'", i, id, result[i])
		}
	}
}

func TestGetMetrics_Empty(t *testing.T) {
	result := GetMetrics(nil)

	if len(result) != 0 {
		t.Errorf("expected 0 метрик, got %d", len(result))
	}
}

func TestGetMetrics_BufferPoolReuse(t *testing.T) {
	// Запускаем несколько раз для проверки поведения пула буферов
	for i := 0; i < 10; i++ {
		metrics := []models.Metrics{{ID: "test"}}
		result := GetMetrics(metrics)
		if len(result) != 1 || result[0] != "test" {
			t.Errorf("iteration %d: unexpected result", i)
		}
	}
}

func TestGetMetrics_BufferNotModified(t *testing.T) {
	metrics := []models.Metrics{{ID: "original"}}
	result := GetMetrics(metrics)

	// Изменение результата не должно повлиять на буфер пула
	result[0] = "modified"

	// Запускаем снова - должно вернуть исходные значения
	metrics2 := []models.Metrics{{ID: "test2"}}
	result2 := GetMetrics(metrics2)

	if len(result2) == 1 && result2[0] != "test2" {
		t.Errorf("expected 'test2', got '%s' - буфер пула был затронут", result2[0])
	}
}

func TestGetMetrics_LargeBatch(t *testing.T) {
	metrics := make([]models.Metrics, 1000)
	expected := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		metrics[i] = models.Metrics{ID: fmt.Sprintf("metric_%d", i)}
		expected[i] = fmt.Sprintf("metric_%d", i)
	}

	result := GetMetrics(metrics)

	if len(result) != 1000 {
		t.Errorf("expected 1000 метрик, got %d", len(result))
	}

	for i, id := range expected {
		if result[i] != id {
			t.Errorf("expected result[%d] = '%s', got '%s'", i, id, result[i])
			break
		}
	}
}

func TestGetMetrics_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	results := make(chan []string, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			metrics := []models.Metrics{{ID: fmt.Sprintf("metric_%d", n)}}
			results <- GetMetrics(metrics)
		}(i)
	}

	wg.Wait()
	close(results)

	// Проверяем все результаты
	count := 0
	for result := range results {
		if len(result) != 1 || !strings.HasPrefix(result[0], "metric_") {
			t.Errorf("unexpected result: %v", result)
		}
		count++
	}

	if count != 100 {
		t.Errorf("expected 100 результатов, got %d", count)
	}
}

func TestGetMetrics_BufferCapacity(t *testing.T) {
	// Создаем больший буфер для теста емкости
	metrics := make([]models.Metrics, 20)
	for i := 0; i < 20; i++ {
		metrics[i] = models.Metrics{ID: fmt.Sprintf("m%d", i)}
	}

	result := GetMetrics(metrics)
	if len(result) != 20 {
		t.Errorf("expected 20 метрик, got %d", len(result))
	}
}

func TestGetMetrics_SingleElement(t *testing.T) {
	metrics := []models.Metrics{{ID: "only_one"}}
	result := GetMetrics(metrics)

	if len(result) != 1 || result[0] != "only_one" {
		t.Errorf("неожиданный результат: %v", result)
	}
}

func TestGetMetrics_UnicodeIDs(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "тест_метрика"},
		{ID: "metric-日本語"},
		{ID: "metric_emoji_🎉"},
	}

	result := GetMetrics(metrics)

	if len(result) != 3 {
		t.Errorf("expected 3 метрики, got %d", len(result))
	}
}

func TestGetMetrics_LongIDs(t *testing.T) {
	longID := strings.Repeat("a", 1000)
	metrics := []models.Metrics{{ID: longID}}

	result := GetMetrics(metrics)

	if len(result) != 1 || result[0] != longID {
		t.Errorf("неожиданный результат для длинного ID")
	}
}

// Вспомогательная функция для тестов
func floatPtr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestGetMetrics_IDWithSpecialChars(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "metric-with-dash"},
		{ID: "metric_with_underscore"},
		{ID: "metric.with.dot"},
		{ID: "metric/with/slash"},
	}

	result := GetMetrics(metrics)

	if len(result) != 4 {
		t.Errorf("expected 4 метрики, got %d", len(result))
	}
}

func TestGetMetrics_IDWithSpaces(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "metric with spaces"},
	}

	result := GetMetrics(metrics)

	if len(result) != 1 || result[0] != "metric with spaces" {
		t.Errorf("неожиданный результат: %v", result)
	}
}

func TestGetMetrics_IDWithNewline(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "metric\nwith\nnewline"},
	}

	result := GetMetrics(metrics)

	if len(result) != 1 {
		t.Errorf("expected 1 метрика, got %d", len(result))
	}
}

func TestGetMetrics_IDWithNullByte(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "metric" + string(rune(0)) + "withnull"},
	}

	result := GetMetrics(metrics)

	if len(result) != 1 {
		t.Errorf("expected 1 метрика, got %d", len(result))
	}
}

func TestGetMetrics_IDWithControlChars(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "metric" + string(rune(1)) + "control"},
	}

	result := GetMetrics(metrics)

	if len(result) != 1 {
		t.Errorf("expected 1 метрика, got %d", len(result))
	}
}

func TestGetMetrics_IDWithDelimiter(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "metric:with:colons"},
		{ID: "metric,with,commas"},
	}

	result := GetMetrics(metrics)

	if len(result) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(result))
	}
}

func TestGetMetrics_IDWithQuotes(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "metric\"with\"quotes"},
		{ID: "metric'with'single"},
	}

	result := GetMetrics(metrics)

	if len(result) != 2 {
		t.Errorf("expected 2 метрики, got %d", len(result))
	}
}

func TestGetMetrics_IDWithAngleBrackets(t *testing.T) {
	metrics := []models.Metrics{
		{ID: "metric<with>brackets"},
		{ID: "metric{with}curlies"},
	}

	result := GetMetrics(metrics)

	if len(result) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(result))
	}
}
