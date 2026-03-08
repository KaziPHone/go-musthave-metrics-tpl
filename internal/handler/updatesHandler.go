package handlers

import (
	"net/http"

	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/helpers"
)

func (h *Handler) UpdatesHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
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
