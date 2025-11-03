package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Storage storage.IStorage
}

func (h *Handler) UpdateValueHandler(w http.ResponseWriter, r *http.Request) {

	typeMetric := chi.URLParam(r, "typeMetric")
	nameMetric := chi.URLParam(r, "nameMetric")
	valueStr := chi.URLParam(r, "value")

	if nameMetric == "" || valueStr == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var value interface{}
	var err error

	switch typeMetric {
	case "gauge":
		value, err = strconv.ParseFloat(valueStr, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case "counter":
		value, err = strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Storage.UpdateMetric(nameMetric, typeMetric, value)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Updated successfully")
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	metric, err := h.readerMetricRequest(r)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if metric.ID == "" || metric.MType == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = h.Storage.UpdateMetricV2(*metric)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}
