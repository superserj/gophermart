package handler

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// Withdrawals — GET /api/user/withdrawals.
func (h *Handler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	list, err := h.store.ListWithdrawals(r.Context(), userID)
	if err != nil {
		h.log.Warn("list withdrawals failed", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(list)
}
