package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/KaziPHone/go-musthave-metrics-tpl/cmd/server/storage"
)

type Handler struct {
	Storage storage.MemStorage
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")

	if path[1] != "update" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(path) != 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	typeMetric := path[2]
	metricName := path[3]
	valueStr := path[4]

	if metricName == "" || valueStr == "" {
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

	h.Storage.UpdateMetric(metricName, typeMetric, value)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Updated successfully")
}
