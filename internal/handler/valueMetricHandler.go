package handlers

import (
	"encoding/json"

	"net/http"
)

func (h *Handler) ValueMetricHandler(w http.ResponseWriter, r *http.Request) {
	metric, err := h.reader(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if v, ok := h.Storage.GetMetric(metric.ID); !ok {

	} else {
		if metric.MType == "gauge" {
			metric.Value = &v.Gauge
		} else {
			*metric.Value = float64(v.Counter)
		}
		metric.MType = "gauge"
		resp, err := json.MarshalIndent(metric, "", " ")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}

}
