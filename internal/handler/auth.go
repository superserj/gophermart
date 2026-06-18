package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/superserj/gophermart/internal/auth"
	"github.com/superserj/gophermart/internal/logger"
	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/repository"
)

// maxJSONBodyBytes — щедрый лимит на тело JSON-запроса. Реальные {login,password}/
// {order,sum} крошечные; лимит защищает от memory-DoS и gzip-бомбы.
const maxJSONBodyBytes = 1 << 16 // 64 КБ

// Register — POST /api/user/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.AuthRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		logger.Log.Warn("hash password failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	id, err := h.store.CreateUser(r.Context(), req.Login, hash)
	if err != nil {
		if errors.Is(err, repository.ErrLoginTaken) {
			http.Error(w, "login already taken", http.StatusConflict)
			return
		}
		logger.Log.Warn("create user failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h.auth.SetAuthCookie(w, strconv.FormatInt(id, 10))
	w.WriteHeader(http.StatusOK)
}

// Login — POST /api/user/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.AuthRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password are required", http.StatusBadRequest)
		return
	}
	id, hash, err := h.store.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "invalid login or password", http.StatusUnauthorized)
			return
		}
		logger.Log.Warn("get user failed", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := auth.CheckPassword(hash, req.Password); err != nil {
		http.Error(w, "invalid login or password", http.StatusUnauthorized)
		return
	}
	h.auth.SetAuthCookie(w, strconv.FormatInt(id, 10))
	w.WriteHeader(http.StatusOK)
}
