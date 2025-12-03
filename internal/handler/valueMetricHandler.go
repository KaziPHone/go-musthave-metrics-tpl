package handlers

import (
	"encoding/json"

	"net/http"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

func (h *Handler) ValueMetricHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	metric, err := h.singleMetric(r)

	mResponse := models.Metrics{
		ID:    metric.ID,
		MType: metric.MType,
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if v, ok := h.Storage.GetMetric(metric.ID); !ok {
		w.WriteHeader(http.StatusNotFound)
		resp, err := json.MarshalIndent(mResponse, "", " ")
		if err != nil {
			w.Write([]byte(err.Error()))
		} else {
			w.Write(resp)
		}
		return
	} else {

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
	}

}
