package agent

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	cfg "github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

func TestSendMetric_HTTP(t *testing.T) {
	// тестовый сервер принимает сжатые данные и проверяет заголовки
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Fatalf("expected Content-Encoding gzip, got %s", r.Header.Get("Content-Encoding"))
		}
		var reader io.Reader = r.Body
		gr, err := gzip.NewReader(r.Body)
		if err == nil {
			defer gr.Close()
			reader = gr
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var metrics []models.Metrics
		if err := json.Unmarshal(data, &metrics); err != nil {
			t.Fatalf("unmarshal metrics: %v", err)
		}
		if len(metrics) == 0 {
			t.Fatalf("no metrics received")
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	addr := srv.Listener.Addr().String()
	a := NewAgent(cfg.AgentConfig{Host: addr, PollInterval: 1, ReportInterval: 1, RateLimit: 1})
	a.url = srv.URL + "/updates/"

	metrics := []models.Metrics{{ID: "m_test", MType: models.Gauge, Value: func() *float64 { v := 9.9; return &v }()}}

	a.sendMetric(metrics)
}
