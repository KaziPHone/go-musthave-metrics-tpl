// Package handlers предоставляет HTTP-обработчики для метрик.
//
// Обработчики реализуют REST API для работы с метриками:
//   - GET /value/{type}/{name} — получить значение метрики
//   - GET /list — получить список всех метрик
//   - POST /update/{type}/{name} — обновить одну метрику
//   - POST /update — обновить одну метрику из JSON
//   - POST /updates — обновить несколько метрик
//
// Поддерживает сжатие gzip и хэширование SHA256 через middleware.
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/KaziPHone/go-musthave-metrics-tpl/internal/audit"
	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
)

// Handler обработчик HTTP-запросов к метрикам.
//
// Содержит ссылки на хранилище метрик и предмет аудита для отправки
// уведомлений о изменениях.
type Handler struct {
	Storage      storage.IStorage
	AuditSubject *audit.Subject
}

// singleMetric декодирует JSON тело запроса в одну метрику.
//
// Параметры:
//   - r: HTTP запрос
//
// Возвращает:
//   - *models.Metrics: декодированная метрика
//   - error: ошибка декодирования или nil
func (h *Handler) singleMetric(r *http.Request) (*models.Metrics, error) {
	var metric models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		return nil, err
	}
	return &metric, nil
}

// multipleMetrics декодирует JSON тело запроса в массив метрик.
//
// Параметры:
//   - r: HTTP запрос
//
// Возвращает:
//   - []models.Metrics: декодированные метрики
//   - error: ошибка декодирования или nil
func (h *Handler) multipleMetrics(r *http.Request) ([]models.Metrics, error) {
	var metrics []models.Metrics

	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}

// NotifyAudit отправляет событие аудита о изменениях в метриках.
//
// Параметры:
//   - metrics: список измененных метрик
//   - ip: IP-адрес клиента
//
// Создает AuditEvent и уведомляет всех наблюдателей через Subject.
func (h *Handler) NotifyAudit(metrics []string, ip string) {
	if h.AuditSubject != nil {
		// Reuse a single slice for the event to reduce allocations
		event := audit.AuditEvent{
			TS:        time.Now().Unix(),
			Metrics:   metrics,
			IPAddress: ip,
		}
		h.AuditSubject.Notify(event)
	}
}
