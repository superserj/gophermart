package handler

import "net/http"

// Ping — GET /ping: 200 если БД доступна, иначе 500.
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(r.Context()); err != nil {
		http.Error(w, "database unavailable", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
