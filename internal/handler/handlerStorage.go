package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/audit"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
)

type Handler struct {
	Storage      storage.IStorage
	AuditSubject *audit.Subject
}

func (h *Handler) singleMetric(r *http.Request) (*models.Metrics, error) {

	var metric models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		return nil, err
	}
	return &metric, nil
}

func (h *Handler) multipleMetrics(r *http.Request) ([]models.Metrics, error) {

	var metrics []models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}

func (h *Handler) NotifyAudit(metrics []string, ip string) {
	if h.AuditSubject != nil {

		event := audit.AuditEvent{
			Ts:        time.Now().Unix(),
			Metrics:   metrics,
			IPAddress: ip,
		}
		h.AuditSubject.Notify(event)
	}
}
