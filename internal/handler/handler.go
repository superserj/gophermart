package handler

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/superserj/gophermart/internal/auth"
	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
)

// Store — зависимости HTTP-слоя. Интерфейс растёт по инкрементам.
// Идентификатор пользователя передаётся строкой: HTTP-слой обращается с ним как
// с непрозрачным токеном, а формат хранения известен только репозиторию.
type Store interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, login, passwordHash string) (string, error)
	GetUserByLogin(ctx context.Context, login string) (id, passwordHash string, err error)
	SaveOrder(ctx context.Context, number, userID string) (bool, error)
	ListOrdersByUser(ctx context.Context, userID string) ([]model.Order, error)
	GetBalance(ctx context.Context, userID string) (current, withdrawn money.Points, err error)
	Withdraw(ctx context.Context, userID, order string, sum money.Points) error
	ListWithdrawals(ctx context.Context, userID string) ([]model.Withdrawal, error)
}

// Handler держит зависимости HTTP-слоя.
type Handler struct {
	store Store
	auth  *auth.Authenticator
	log   *zap.Logger
}

// New собирает Handler. Логгер передаётся явно, без обращения к глобальному состоянию.
func New(store Store, a *auth.Authenticator, log *zap.Logger) *Handler {
	return &Handler{store: store, auth: a, log: log}
}

// currentUserID извлекает идентификатор пользователя из контекста (после
// RequireAuth) как непрозрачную строку, не делая предположений о его формате.
func currentUserID(r *http.Request) (string, bool) {
	raw, ok := auth.UserIDFromContext(r.Context())
	if !ok || raw == "" {
		return "", false
	}
	return raw, true
}
