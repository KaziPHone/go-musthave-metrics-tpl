package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetMetricHandler(w http.ResponseWriter, r *http.Request) {

	typeMetric := chi.URLParam(r, "typeMetric")
	nameMetric := chi.URLParam(r, "nameMetric")

	metric, ok := h.Storage.MetricTypes[nameMetric]

	if !ok || !(typeMetric == "gauge" || typeMetric == "counter") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if typeMetric == "gauge" {
		fmt.Fprintln(w, strconv.FormatFloat(metric.Gauge, 'f', -1, 64))
	} else {
		fmt.Fprintln(w, strconv.FormatInt(metric.Counter, 10))
	}
}
