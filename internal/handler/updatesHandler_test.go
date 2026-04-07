package handlers

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http/httptest"
	"testing"

	cfg "github.com/KaziPHone/go-musthave-metrics-tpl/internal/config"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
)

func TestUpdatesHandler_Forbidden(t *testing.T) {
	// создаём доверенную подсеть 192.168.0.0/24
	_, ipnet, err := net.ParseCIDR("192.168.0.0/24")
	if err != nil {
		t.Fatalf("ParseCIDR failed: %v", err)
	}

	serverCfg := cfg.ServerConfig{}
	st := storage.NewMemStorage(serverCfg)

	h := &Handler{Storage: st, TrustedSubnet: ipnet}

	metrics := []models.Metrics{{ID: "m1", MType: models.Gauge, Value: func() *float64 { v := 1.23; return &v }()}}
	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-IP", "10.0.0.5")

	rr := httptest.NewRecorder()
	h.UpdatesHandler(rr, req)

	if rr.Code != 403 {
		t.Fatalf("expected 403, got %d, body=%s", rr.Code, rr.Body.String())
	}
}

func TestUpdatesHandler_Allowed(t *testing.T) {
	_, ipnet, err := net.ParseCIDR("192.168.0.0/24")
	if err != nil {
		t.Fatalf("ParseCIDR failed: %v", err)
	}

	serverCfg := cfg.ServerConfig{}
	st := storage.NewMemStorage(serverCfg)

	h := &Handler{Storage: st, TrustedSubnet: ipnet}

	metrics := []models.Metrics{{ID: "m_ok", MType: models.Gauge, Value: func() *float64 { v := 3.21; return &v }()}}
	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest("POST", "/updates/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-IP", "192.168.0.5")

	rr := httptest.NewRecorder()
	h.UpdatesHandler(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d, body=%s", rr.Code, rr.Body.String())
	}

	if mt, ok := st.GetMetric("m_ok"); !ok || mt == nil {
		t.Fatalf("expected metric m_ok to be stored")
	}
}
