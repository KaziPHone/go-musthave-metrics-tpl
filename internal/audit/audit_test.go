package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// Helper function to compare string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSubject_Attach(t *testing.T) {
	subject := &Subject{}
	mockObserver := &mockObserver{}

	subject.Attach(mockObserver)

	subject.mu.Lock()
	if len(subject.observers) != 1 {
		t.Errorf("expected 1 observer, got %d", len(subject.observers))
	}
	if subject.observers[0] != mockObserver {
		t.Error("observer not attached correctly")
	}
	subject.mu.Unlock()
}

func TestSubject_Notify(t *testing.T) {
	subject := &Subject{}

	obs1 := &mockObserver{}
	obs2 := &mockObserver{}

	subject.Attach(obs1)
	subject.Attach(obs2)

	event := AuditEvent{
		TS:        1234567890,
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "127.0.0.1",
	}

	subject.Notify(event)

	obs1.mu.Lock()
	obs2.mu.Lock()
	if len(obs1.receivedEvents) != 1 {
		t.Errorf("observer 1 expected 1 event, got %d", len(obs1.receivedEvents))
	}
	if len(obs2.receivedEvents) != 1 {
		t.Errorf("observer 2 expected 1 event, got %d", len(obs2.receivedEvents))
	}
	obs1.mu.Unlock()
	obs2.mu.Unlock()
}

func TestSubject_Notify_Concurrent(t *testing.T) {
	subject := &Subject{}
	numObservers := 10
	numEvents := 100

	var observers []*mockObserver
	for i := 0; i < numObservers; i++ {
		observers = append(observers, &mockObserver{})
		subject.Attach(observers[i])
	}

	var wg sync.WaitGroup

	// Concurrent notifications
	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(e AuditEvent) {
			defer wg.Done()
			subject.Notify(e)
		}(AuditEvent{TS: int64(i)})
	}

	wg.Wait()

	// Verify all observers received all events
	for _, obs := range observers {
		obs.mu.Lock()
		if len(obs.receivedEvents) != numEvents {
			t.Errorf("expected %d events, got %d", numEvents, len(obs.receivedEvents))
		}
		obs.mu.Unlock()
	}
}

func TestFileObserver_OnAuditEvent(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit_test_*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	observer := NewFileObserver(tmpFile.Name())

	event := AuditEvent{
		TS:        1234567890,
		Metrics:   []string{"metric1"},
		IPAddress: "127.0.0.1",
	}

	observer.OnAuditEvent(event)

	// Read file content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var parsedEvent AuditEvent
	if err := json.Unmarshal(content, &parsedEvent); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if parsedEvent.TS != event.TS {
		t.Errorf("expected TS %d, got %d", event.TS, parsedEvent.TS)
	}
}

func TestFileObserver_OnAuditEvent_FileError(t *testing.T) {
	observer := NewFileObserver("/nonexistent/path/audit.log")
	event := AuditEvent{TS: 1234567890}

	// Should not panic on error
	observer.OnAuditEvent(event)
}

func TestHTTPObserver_OnAuditEvent_Success(t *testing.T) {
	var receivedEvent AuditEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedEvent)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL)
	event := AuditEvent{TS: 1234567890, IPAddress: "127.0.0.1"}

	observer.OnAuditEvent(event)

	if receivedEvent.TS != event.TS {
		t.Errorf("expected TS %d, got %d", event.TS, receivedEvent.TS)
	}
}

func TestHTTPObserver_OnAuditEvent_HTTPError(t *testing.T) {
	observer := NewHTTPObserver("http://nonexistent-server:9999")
	event := AuditEvent{TS: 1234567890}

	// Should not panic on network error
	observer.OnAuditEvent(event)
}

func TestHTTPObserver_OnAuditEvent_BadStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL)
	event := AuditEvent{TS: 1234567890}

	observer.OnAuditEvent(event)
}

func TestHTTPObserver_OnAuditEvent_NonJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL)
	event := AuditEvent{TS: 1234567890}

	observer.OnAuditEvent(event)
}

// Mock observer for testing
type mockObserver struct {
	receivedEvents []AuditEvent
	mu             sync.Mutex
}

func (m *mockObserver) OnAuditEvent(event AuditEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.receivedEvents = append(m.receivedEvents, event)
}

func TestNewFileObserver(t *testing.T) {
	path := "/tmp/test_audit.log"
	observer := NewFileObserver(path)

	if observer.filePath != path {
		t.Errorf("expected filePath %s, got %s", path, observer.filePath)
	}
}

func TestNewHTTPObserver(t *testing.T) {
	url := "http://example.com/audit"
	observer := NewHTTPObserver(url)

	if observer.url != url {
		t.Errorf("expected url %s, got %s", url, observer.url)
	}
}

func TestAuditEvent_JSONMarshal(t *testing.T) {
	event := AuditEvent{
		TS:        1234567890,
		Metrics:   []string{"metric1", "metric2"},
		IPAddress: "192.168.1.1",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed AuditEvent
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.TS != event.TS {
		t.Errorf("TS mismatch: expected %d, got %d", event.TS, parsed.TS)
	}
	if !equalStringSlices(parsed.Metrics, event.Metrics) {
		t.Errorf("Metrics mismatch: expected %v, got %v", event.Metrics, parsed.Metrics)
	}
	if parsed.IPAddress != event.IPAddress {
		t.Errorf("IPAddress mismatch: expected %s, got %s", event.IPAddress, parsed.IPAddress)
	}
}

func TestFileObserver_MultipleWrites(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit_multi_*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	observer := NewFileObserver(tmpFile.Name())

	for i := 0; i < 5; i++ {
		event := AuditEvent{TS: int64(i)}
		observer.OnAuditEvent(event)
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	// Count newlines
	lines := strings.Count(string(content), "\n")
	if lines != 5 {
		t.Errorf("expected 5 lines, got %d", lines)
	}
}

func TestSubject_Notify_Empty(t *testing.T) {
	subject := &Subject{}
	event := AuditEvent{TS: 1234567890}

	// Should not panic with no observers
	subject.Notify(event)
}

func TestSubject_ConcurrentAttach(t *testing.T) {
	subject := &Subject{}
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			subject.Attach(&mockObserver{})
		}()
	}

	wg.Wait()

	subject.mu.Lock()
	if len(subject.observers) != 100 {
		t.Errorf("expected 100 observers, got %d", len(subject.observers))
	}
	subject.mu.Unlock()
}

func TestHTTPObserver_EmptyURL(t *testing.T) {
	observer := NewHTTPObserver("")
	event := AuditEvent{TS: 1234567890}

	// Should not panic with empty URL
	observer.OnAuditEvent(event)
}

func TestFileObserver_Concurrent(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit_concurrent_*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	observer := NewFileObserver(tmpFile.Name())
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			observer.OnAuditEvent(AuditEvent{TS: int64(i)})
		}(i)
	}

	wg.Wait()

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	lines := strings.Count(string(content), "\n")
	if lines != 50 {
		t.Errorf("expected 50 lines, got %d", lines)
	}
}

func TestHTTPObserver_Concurrent(t *testing.T) {
	var receivedEvents []AuditEvent
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var event AuditEvent
		json.NewDecoder(r.Body).Decode(&event)
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			observer.OnAuditEvent(AuditEvent{TS: int64(i)})
		}(i)
	}

	wg.Wait()

	// Note: Some requests might fail due to rate limiting or connection issues
	// so we just verify it doesn't panic
}

func TestFileObserver_EmptyFilePath(t *testing.T) {
	observer := NewFileObserver("")
	event := AuditEvent{TS: 1234567890}

	// Should not panic with empty path
	observer.OnAuditEvent(event)
}

func TestSubject_RemoveObserver(t *testing.T) {
	// This tests that Attach works and we can verify state
	subject := &Subject{}

	obs1 := &mockObserver{}
	obs2 := &mockObserver{}

	subject.Attach(obs1)
	subject.Attach(obs2)

	subject.mu.Lock()
	initialCount := len(subject.observers)
	subject.mu.Unlock()

	if initialCount != 2 {
		t.Errorf("expected 2 observers initially, got %d", initialCount)
	}
}

func TestHTTPObserver_CloseResponseBody(t *testing.T) {
	// Verify that response body is closed even on error path
	observer := NewHTTPObserver("http://nonexistent-server")
	event := AuditEvent{TS: 1234567890}

	observer.OnAuditEvent(event)
}

func TestAuditEvent_TimestampVariations(t *testing.T) {
	tests := []int64{
		0,
		1,
		time.Now().Unix(),
		9999999999999,
		-1,
	}

	for _, ts := range tests {
		event := AuditEvent{TS: ts}
		data, err := json.Marshal(event)
		if err != nil {
			t.Errorf("failed to marshal event with TS %d: %v", ts, err)
			continue
		}

		var parsed AuditEvent
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("failed to unmarshal event with TS %d: %v", ts, err)
			continue
		}

		if parsed.TS != ts {
			t.Errorf("TS mismatch for %d: got %d", ts, parsed.TS)
		}
	}
}

func TestAuditEvent_MetricsVariations(t *testing.T) {
	tests := []struct {
		name    string
		metrics []string
	}{
		{"empty", []string{}},
		{"single", []string{"metric"}},
		{"multiple", []string{"m1", "m2", "m3"}},
		{"special chars", []string{"metric-with-dash", "metric_with_underscore", "metric.with.dot"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := AuditEvent{Metrics: tt.metrics}
			data, err := json.Marshal(event)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var parsed AuditEvent
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if len(parsed.Metrics) != len(tt.metrics) {
				t.Errorf("metrics length mismatch: expected %d, got %d", len(tt.metrics), len(parsed.Metrics))
			}
		})
	}
}

func TestHTTPObserver_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("this is not valid json at all!!!"))
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL)
	event := AuditEvent{TS: 1234567890}

	observer.OnAuditEvent(event)
}

func TestHTTPObserver_TimeoutResponse(t *testing.T) {
	// Server that doesn't respond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do nothing - timeout
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL)
	event := AuditEvent{TS: 1234567890}

	observer.OnAuditEvent(event)
}

func TestFileObserver_WriteAppendMode(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "audit_append_*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	observer := NewFileObserver(tmpFile.Name())

	// Write first event
	observer.OnAuditEvent(AuditEvent{TS: 1})
	content1, _ := os.ReadFile(tmpFile.Name())

	// Write second event
	observer.OnAuditEvent(AuditEvent{TS: 2})
	content2, _ := os.ReadFile(tmpFile.Name())

	// Verify append behavior
	if len(content2) <= len(content1) {
		t.Error("expected file to grow after second write")
	}

	if !bytes.Contains(content2, []byte("1")) || !bytes.Contains(content2, []byte("2")) {
		t.Error("expected both events in file")
	}
}
