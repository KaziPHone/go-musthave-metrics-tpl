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

	metric, ok := h.Storage.MetricTypes[typeMetric]

	if !ok || !(nameMetric == "gauge" || nameMetric == "counter") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if nameMetric == "gauge" {
		fmt.Fprintf(w, "%s: Gauge=%v", typeMetric, strconv.FormatFloat(metric.Gauge, 'f', -1, 64))
	} else {
		fmt.Fprintf(w, "%s: Counter=%v", typeMetric, strconv.FormatInt(metric.Counter, 10))
	}
}
