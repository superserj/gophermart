package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
)

func TestWithdrawalsOK(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:57+03:00")
	h := handlerWith(&fakeStore{list: []model.Withdrawal{{Order: "2377225624", Sum: money.FromFloat(500), ProcessedAt: ts}}})
	rr := httptest.NewRecorder()
	h.Withdrawals(rr, authedBal(http.MethodGet, "/api/user/withdrawals", ""))
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	require.JSONEq(t, `[{"order":"2377225624","sum":500.00,"processed_at":"2020-12-09T16:09:57+03:00"}]`, rr.Body.String())
}

func TestWithdrawalsNoContent(t *testing.T) {
	h := handlerWith(&fakeStore{list: nil})
	rr := httptest.NewRecorder()
	h.Withdrawals(rr, authedBal(http.MethodGet, "/api/user/withdrawals", ""))
	require.Equal(t, http.StatusNoContent, rr.Code)
}

func TestWithdrawalsUnauthorized(t *testing.T) {
	h := handlerWith(&fakeStore{})
	rr := httptest.NewRecorder()
	h.Withdrawals(rr, httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil))
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestWithdrawalsInternalError(t *testing.T) {
	h := handlerWith(&fakeStore{listErr: context.DeadlineExceeded})
	rr := httptest.NewRecorder()
	h.Withdrawals(rr, authedBal(http.MethodGet, "/api/user/withdrawals", ""))
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}
