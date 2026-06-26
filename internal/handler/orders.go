package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"encoding/json"

	"go.uber.org/zap"

	"github.com/superserj/gophermart/internal/luhn"
	"github.com/superserj/gophermart/internal/repository"
)

// maxOrderBodyBytes — щедрый лимит тела POST /api/user/orders: номер заказа
// короткий, реальный клиент/автотест укладывается, мусорный payload отсекается.
const maxOrderBodyBytes = 4096

// UploadOrder — POST /api/user/orders. Тело — голый номер заказа (text/plain).
func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	// Номер заказа короткий; ограничиваем тело, чтобы мусорный payload
	// не прошёл Луна и не попал в TEXT-ключ (защита от DoS).
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxOrderBodyBytes))
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
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
		h.log.Warn("save order failed", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		h.log.Warn("list orders failed", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
