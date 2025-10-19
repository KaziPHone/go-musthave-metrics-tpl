package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handler) ListMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Список текущих метрик:</h1>")
	for key, val := range h.Storage.ListMetrics() {
		fmt.Fprintf(w, "%s: Counter=%v, Gauge=%v<br>", key, val.Counter, val.Gauge)
	}
}
