package handler

import (
	"github.com/go-chi/chi/v5"

	"github.com/superserj/gophermart/internal/auth"
	"github.com/superserj/gophermart/internal/logger"
	"github.com/superserj/gophermart/internal/middleware"
)

// NewRouter собирает HTTP-роутер: глобально логирование и gzip,
// публичные register/login и /ping, защищённая группа под RequireAuth.
func NewRouter(h *Handler, a *auth.Authenticator) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.WithLogging)
	r.Use(middleware.Gzip)

	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)
	r.Get("/ping", h.Ping)

	r.Group(func(pr chi.Router) {
		pr.Use(a.RequireAuth)
		h.registerProtected(pr)
	})
	return r
}

// registerProtected — точка расширения для защищённых маршрутов (orders, balance).
func (h *Handler) registerProtected(r chi.Router) {
	r.Post("/api/user/orders", h.UploadOrder)
	r.Get("/api/user/orders", h.ListOrders)
}
