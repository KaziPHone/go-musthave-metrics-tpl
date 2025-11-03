package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	models "github.com/KaziPHone/go-musthave-metrics-tpl/internal/model"
)

func (h *Handler) reader(r *http.Request) (*models.Metrics, error) {
	var buf bytes.Buffer
	var metric models.Metrics

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		return nil, err
	}
	return &metric, err

}
