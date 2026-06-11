package handler

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/superserj/gophermart/internal/logger"
	"github.com/superserj/gophermart/internal/model"
)

// Balance — GET /api/user/balance. Всегда 200 с телом.
func (h *Handler) Balance(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	current, withdrawn, err := h.store.GetBalance(r.Context(), userID)
	if err != nil {
		logger.Log.Warn("get balance failed", zap.Error(err))
		http.Error(w, "balance failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.Balance{Current: current, Withdrawn: withdrawn})
}
