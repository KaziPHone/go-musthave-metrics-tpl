package handlers

import "net/http"

func (h *Handler) GetPingDBHandler(w http.ResponseWriter, r *http.Request) {
	if !h.Storage.IsConnectedDB() {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
