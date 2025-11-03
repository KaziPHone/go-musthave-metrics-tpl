package handlers

import (
	"encoding/json"

	"net/http"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

func (h *Handler) ValueMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	metric, err := h.readerMetricRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if v, ok := h.Storage.GetMetric(metric.ID); !ok {
		resp, err := json.MarshalIndent(v, "", " ")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(resp)
	} else {
		mResponse := models.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
		}
		if metric.MType == models.Gauge {
			mResponse.Value = &v.Gauge
		} else {
			mResponse.Delta = &v.Counter
		}
		resp, err := json.MarshalIndent(mResponse, "", " ")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(resp)
		w.WriteHeader(http.StatusOK)
	}

}
