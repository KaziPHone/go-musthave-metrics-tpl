package handlers

import (
	"net"
	"net/http"
	"strings"

	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/helpers"
)

// UpdatesHandler обрабатывает пакетную загрузку метрик.
// Перед обработкой проверяет заголовок X-Real-IP на принадлежность доверённой подсети,
// если она задана в конфигурации. В противном случае отдаёт 403 Forbidden.
func (h *Handler) UpdatesHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if h.TrustedSubnet != nil {
		ipStr := r.Header.Get("X-Real-IP")
		if ipStr == "" {
			ipStr = helpers.GetClientIP(r)
		}
		ipStr = strings.TrimSpace(ipStr)
		ip := net.ParseIP(ipStr)
		if ip == nil || !h.TrustedSubnet.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("forbidden"))
			return
		}
	}

	metrics, err := h.multipleMetrics(r)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	h.Storage.UpdatesMetrics(metrics)

	go h.NotifyAudit(helpers.GetMetrics(metrics), helpers.GetClientIP(r))

	w.Write([]byte("{}"))
}
