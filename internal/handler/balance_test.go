package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/superserj/gophermart/internal/auth"
	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
)

func authedBal(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	return req.WithContext(auth.WithUserID(req.Context(), "7"))
}

func TestBalanceOK(t *testing.T) {
	h := &Handler{store: &fakeStore{balCurrent: money.FromFloat(729.98), balWithdrawn: 0}}
	rr := httptest.NewRecorder()
	h.Balance(rr, authedBal(http.MethodGet, "/api/user/balance", ""))
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	var got model.Balance
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
	require.Equal(t, money.FromFloat(729.98), got.Current)
	require.JSONEq(t, `{"current":729.98,"withdrawn":0.00}`, rr.Body.String())
}

func TestBalanceUnauthorized(t *testing.T) {
	h := &Handler{store: &fakeStore{}}
	rr := httptest.NewRecorder()
	h.Balance(rr, httptest.NewRequest(http.MethodGet, "/api/user/balance", nil))
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestBalanceInternalError(t *testing.T) {
	h := &Handler{store: &fakeStore{balErr: context.DeadlineExceeded}}
	rr := httptest.NewRecorder()
	h.Balance(rr, authedBal(http.MethodGet, "/api/user/balance", ""))
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}
