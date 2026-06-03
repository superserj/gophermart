package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"encoding/json"

	"go.uber.org/zap"

	"github.com/superserj/gophermart/internal/logger"
	"github.com/superserj/gophermart/internal/luhn"
	"github.com/superserj/gophermart/internal/repository"
)

// UploadOrder — POST /api/user/orders. Тело — голый номер заказа (text/plain).
func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusInternalServerError)
		return
	}
	number := strings.TrimSpace(string(body))
	if number == "" {
		http.Error(w, "empty order number", http.StatusBadRequest)
		return
	}
	if !luhn.Valid(number) {
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		return
	}
	existed, err := h.store.SaveOrder(r.Context(), number, userID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderOwnedByOther) {
			http.Error(w, "order belongs to another user", http.StatusConflict)
			return
		}
		logger.Log.Warn("save order failed", zap.Error(err))
		http.Error(w, "save failed", http.StatusInternalServerError)
		return
	}
	if existed {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// ListOrders — GET /api/user/orders. Всегда application/json (в т.ч. при 204).
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	orders, err := h.store.ListOrdersByUser(r.Context(), userID)
	if err != nil {
		logger.Log.Warn("list orders failed", zap.Error(err))
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(orders)
}
