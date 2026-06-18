package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/superserj/gophermart/internal/logger"
	"github.com/superserj/gophermart/internal/luhn"
	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
	"github.com/superserj/gophermart/internal/repository"
)

// Withdraw — POST /api/user/balance/withdraw.
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req model.WithdrawRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	sum := money.FromFloat(req.Sum)
	if sum <= 0 || !luhn.Valid(req.Order) {
		http.Error(w, "invalid order or sum", http.StatusUnprocessableEntity)
		return
	}
	err := h.store.Withdraw(r.Context(), userID, req.Order, sum)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, repository.ErrInvalidWithdrawal):
		http.Error(w, "invalid order or sum", http.StatusUnprocessableEntity)
	case errors.Is(err, repository.ErrInsufficientFunds):
		http.Error(w, "insufficient funds", http.StatusPaymentRequired)
	default:
		logger.Log.Warn("withdraw failed", zap.Error(err))
		http.Error(w, "withdraw failed", http.StatusInternalServerError)
	}
}
