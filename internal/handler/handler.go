package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/superserj/gophermart/internal/auth"
)

// Store — зависимости HTTP-слоя. Интерфейс растёт по инкрементам.
type Store interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, login, passwordHash string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (int64, string, error)
	// orders (Task 19): SaveOrder, ListOrdersByUser
	// balance (Task 28): GetBalance, Withdraw, ListWithdrawals
}

// Handler держит зависимости HTTP-слоя.
type Handler struct {
	store Store
	auth  *auth.Authenticator
}

// New собирает Handler.
func New(store Store, a *auth.Authenticator) *Handler {
	return &Handler{store: store, auth: a}
}

// currentUserID извлекает int64-идентификатор из контекста (после RequireAuth).
func currentUserID(r *http.Request) (int64, bool) {
	raw, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}
