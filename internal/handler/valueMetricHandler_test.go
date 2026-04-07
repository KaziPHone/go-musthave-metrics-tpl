package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	cfg "github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
)

func TestValueMetricHandler_NotFound(t *testing.T) {
	serverCfg := cfg.ServerConfig{}
	st := storage.NewMemStorage(serverCfg)
	h := &Handler{Storage: st}

	reqBody := models.Metrics{ID: "nope", MType: models.Gauge}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/value/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ValueMetricHandler(rr, req)
	if !bytes.Contains(rr.Body.Bytes(), []byte("nope")) {
		t.Fatalf("expected response to include metric id, got %s", rr.Body.String())
	}
}

func TestValueMetricHandler_Found(t *testing.T) {
	serverCfg := cfg.ServerConfig{}
	st := storage.NewMemStorage(serverCfg)
	// добавить метрику
	v := 5.5
	st.UpdateMetric("m_found", models.Gauge, v)

	h := &Handler{Storage: st}
	reqBody := models.Metrics{ID: "m_found", MType: models.Gauge}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/value/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ValueMetricHandler(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("m_found")) {
		t.Fatalf("expected response to include metric id, got %s", rr.Body.String())
	}
}
