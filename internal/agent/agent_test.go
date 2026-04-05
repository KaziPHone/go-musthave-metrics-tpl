package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

func TestNewAgent(t *testing.T) {
	cfg := config.AgentConfig{
		Host:           "localhost:8080",
		ReportInterval: 10,
		PollInterval:   5,
		Key:            "test-key",
		RateLimit:      5,
	}

	agent := NewAgent(cfg)

	if agent.reportInterval != 10 {
		t.Errorf("expected reportInterval 10, got %d", agent.reportInterval)
	}
	if agent.pollInterval != 5 {
		t.Errorf("expected pollInterval 5, got %d", agent.pollInterval)
	}
	if agent.key != "test-key" {
		t.Errorf("expected key 'test-key', got '%s'", agent.key)
	}
	if agent.rateLimit != 5 {
		t.Errorf("expected rateLimit 5, got %d", agent.rateLimit)
	}
	if agent.maxRetries != 3 {
		t.Errorf("expected maxRetries 3, got %d", agent.maxRetries)
	}
	if len(agent.retryDelays) != 3 {
		t.Errorf("expected 3 retry delays, got %d", len(agent.retryDelays))
	}
}

func TestCompress(t *testing.T) {
	// Use larger data that compresses well
	data := []byte(`{"test": "data", "numbers": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10], "nested": {"key": "value", "array": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15]}}`)

	compressed, err := compress(data)
	if err != nil {
		t.Fatalf("compress failed: %v", err)
	}

	if len(compressed) == 0 {
		t.Error("compressed data should not be empty")
	}

	if len(compressed) >= len(data) {
		t.Errorf("compressed data should be smaller than original for this test data (original: %d, compressed: %d)", len(data), len(compressed))
	}
}

func TestCompress_EmptyData(t *testing.T) {
	data := []byte{}

	compressed, err := compress(data)
	if err != nil {
		t.Fatalf("compress failed: %v", err)
	}

	if compressed == nil {
		t.Error("compressed data should not be nil")
	}
}

func TestAgent_copyMetrics(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := NewAgent(cfg)

	agent.mu.Lock()
	agent.metrics["test_metric"] = 100.5
	agent.metrics["another_metric"] = 200.75
	agent.pollCount = 5
	agent.mu.Unlock()

	metrics := agent.copyMetrics()

	if len(metrics) != 3 {
		t.Errorf("expected 3 metrics, got %d", len(metrics))
	}

	foundPollCount := false
	for _, m := range metrics {
		if m.ID == "PollCount" {
			foundPollCount = true
			if m.MType != models.Counter {
				t.Errorf("expected PollCount to be counter, got %s", m.MType)
			}
			if m.Delta == nil {
				t.Error("expected delta to be set for counter")
			} else if *m.Delta != 5 {
				t.Errorf("expected delta 5, got %d", *m.Delta)
			}
		}
	}

	if !foundPollCount {
		t.Error("expected to find PollCount metric")
	}
}

func TestAgent_copyMetrics_Empty(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := NewAgent(cfg)

	metrics := agent.copyMetrics()

	if len(metrics) != 1 {
		t.Errorf("expected 1 metric (pollCount), got %d", len(metrics))
	}
}

func TestAgent_copyMetrics_Concurrent(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := NewAgent(cfg)

	agent.mu.Lock()
	for i := 0; i < 100; i++ {
		agent.metrics[fmt.Sprintf("metric_%d", i)] = float64(i)
	}
	agent.pollCount = 10
	agent.mu.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metrics := agent.copyMetrics()
			if len(metrics) == 0 {
				t.Error("metrics should not be empty")
			}
		}()
	}

	wg.Wait()
}

func TestAgent_sendMetric_Success(t *testing.T) {
	var receivedMetrics []models.Metrics

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics

		// Handle gzip-encoded body
		var body io.ReadCloser
		if r.Header.Get("Content-Encoding") == "gzip" {
			var err error
			body, err = gzip.NewReader(r.Body)
			if err != nil {
				t.Errorf("failed to create gzip reader: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer body.Close()
		} else {
			body = r.Body
		}

		if err := json.NewDecoder(body).Decode(&metrics); err != nil {
			t.Errorf("failed to decode: %v", err)
		}
		receivedMetrics = metrics
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.AgentConfig{
		Host: server.URL[len("http://"):],
	}
	agent := NewAgent(cfg)

	testMetrics := []models.Metrics{
		{ID: "test1", MType: models.Gauge, Value: floatPtr(100.0)},
		{ID: "test2", MType: models.Counter, Delta: int64Ptr(10)},
	}

	agent.sendMetric(testMetrics)

	if len(receivedMetrics) != 2 {
		t.Errorf("expected 2 metrics received, got %d", len(receivedMetrics))
	}
}

func TestAgent_sendMetric_RetrySuccess(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	cfg := config.AgentConfig{
		Host: server.URL[len("http://"):],
	}
	agent := NewAgent(cfg)

	testMetrics := []models.Metrics{{ID: "test", MType: models.Gauge, Value: floatPtr(1.0)}}

	agent.sendMetric(testMetrics)

	if attempt < 2 {
		t.Errorf("expected retry attempt, only %d made", attempt)
	}
}

func TestAgent_sendMetric_HTTPError(t *testing.T) {
	// Use a fast server that closes connection immediately without retries
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Just close without writing response
		hijacker, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hijacker.Hijack()
			conn.Close()
		}
	}))
	defer server.Close()

	cfg := config.AgentConfig{
		Host:         server.URL[len("http://"):],
		PollInterval: 1, // Minimal intervals for fast tests
	}
	agent := NewAgent(cfg)
	agent.maxRetries = 1                      // Only 1 retry instead of 3
	agent.retryDelays = []time.Duration{0}    // No delay between retries

	testMetrics := []models.Metrics{{ID: "test", MType: models.Gauge, Value: floatPtr(1.0)}}

	agent.sendMetric(testMetrics)
}

func TestAgent_pollMetrics(t *testing.T) {
	cfg := config.AgentConfig{
		PollInterval:   1,
		ReportInterval: 10,
	}
	agent := NewAgent(cfg)

	stopCh := make(chan struct{})
	go agent.pollMetrics(stopCh)

	time.Sleep(1100 * time.Millisecond)

	close(stopCh)

	agent.mu.Lock()
	hasMetrics := len(agent.metrics) > 0
	agent.mu.Unlock()

	if !hasMetrics {
		t.Error("expected metrics to be collected")
	}
}

func TestAgent_reportMetrics(t *testing.T) {
	cfg := config.AgentConfig{
		PollInterval:   1,
		ReportInterval: 1,
		RateLimit:      2,
	}
	agent := NewAgent(cfg)

	agent.mu.Lock()
	agent.metrics["test"] = 100.0
	agent.mu.Unlock()

	stopCh := make(chan struct{})
	go agent.reportMetrics(stopCh)

	time.Sleep(1100 * time.Millisecond)

	close(stopCh)
}

func TestAgent_Start(t *testing.T) {
	cfg := config.AgentConfig{
		PollInterval:   1,
		ReportInterval: 1,
		RateLimit:      1,
	}
	agent := NewAgent(cfg)

	stopCh := make(chan struct{})

	go func() {
		time.Sleep(1100 * time.Millisecond)
		close(stopCh)
	}()

	agent.Start(stopCh)
}

func TestAgent_worker(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := NewAgent(cfg)

	jobs := make(chan []models.Metrics, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go agent.worker(jobs, &wg)

	jobs <- []models.Metrics{{ID: "test", MType: models.Gauge, Value: floatPtr(1.0)}}
	close(jobs)

	wg.Wait()
}

func TestAgent_collectSystemMetrics(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := NewAgent(cfg)

	for i := 0; i < 3; i++ {
		agent.collectSystemMetrics()
	}

	agent.mu.Lock()
	defer agent.mu.Unlock()

	if len(agent.metrics) == 0 {
		t.Error("expected metrics after collection")
	}
}

func TestAgent_collectRuntimeMetrics(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := NewAgent(cfg)

	agent.collectRuntimeMetrics()

	agent.mu.Lock()
	defer agent.mu.Unlock()

	hasRuntimeMetrics := false
	for key := range agent.metrics {
		if bytes.Contains([]byte(key), []byte("Alloc")) ||
			bytes.Contains([]byte(key), []byte("Heap")) ||
			bytes.Contains([]byte(key), []byte("NumGC")) {
			hasRuntimeMetrics = true
			break
		}
	}

	if !hasRuntimeMetrics {
		t.Error("expected runtime metrics")
	}
}

func TestAgent_collectRuntimeMetrics_MultipleCalls(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := NewAgent(cfg)

	for i := 0; i < 5; i++ {
		agent.collectRuntimeMetrics()
	}

	agent.mu.Lock()
	defer agent.mu.Unlock()

	if agent.pollCount != 5 {
		t.Errorf("expected pollCount 5, got %d", agent.pollCount)
	}
}

func TestAgent_concurrentAccess(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := NewAgent(cfg)

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			agent.mu.Lock()
			agent.metrics[fmt.Sprintf("metric_%d", i)] = float64(i)
			agent.mu.Unlock()
		}(i)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			agent.copyMetrics()
		}()
	}

	wg.Wait()
}

func TestAgent_MetricsMapIsolation(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := NewAgent(cfg)

	initial := agent.copyMetrics()

	agent.mu.Lock()
	agent.metrics["new_metric"] = 999.0
	agent.mu.Unlock()

	newMetrics := agent.copyMetrics()

	if len(newMetrics) <= len(initial) {
		t.Error("expected new metrics to include additional metric")
	}
}

func TestAgent_MetricsTypeCounts(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := NewAgent(cfg)

	agent.mu.Lock()
	agent.metrics["gauge1"] = 100.0
	agent.metrics["gauge2"] = 200.0
	agent.metrics["counter1"] = 300.0
	agent.mu.Unlock()

	metrics := agent.copyMetrics()

	gaugeCount := 0
	counterCount := 0
	for _, m := range metrics {
		if m.MType == models.Gauge {
			gaugeCount++
		} else if m.MType == models.Counter {
			counterCount++
		}
	}

	if gaugeCount != 3 {
		t.Errorf("expected 3 gauge metrics, got %d", gaugeCount)
	}
	if counterCount != 1 {
		t.Errorf("expected 1 counter metric (pollCount), got %d", counterCount)
	}
}

func TestAgent_EmptyKey(t *testing.T) {
	cfg := config.AgentConfig{Key: ""}
	agent := NewAgent(cfg)

	if agent.key != "" {
		t.Errorf("expected empty key, got '%s'", agent.key)
	}
}

func TestAgent_retryDelays(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := NewAgent(cfg)

	expectedDelays := []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

	for i, expected := range expectedDelays {
		if agent.retryDelays[i] != expected {
			t.Errorf("retryDelays[%d]: expected %v, got %v", i, expected, agent.retryDelays[i])
		}
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
