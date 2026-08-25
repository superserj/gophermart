package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/superserj/gophermart/internal/money"
	"github.com/superserj/gophermart/internal/repository"
)

func TestWithdrawHandlerBodyTooLarge(t *testing.T) {
	f := &fakeStore{}
	h := handlerWith(f)
	body := `{"order":"` + strings.Repeat("7", maxJSONBodyBytes+1) + `","sum":10}`
	rr := httptest.NewRecorder()
	h.Withdraw(rr, authedBal(http.MethodPost, "/api/user/balance/withdraw", body))
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Empty(t, f.gotOrder, "store must not be called when body exceeds limit")
}

func TestWithdrawHandlerOK(t *testing.T) {
	f := &fakeStore{}
	h := handlerWith(f)
	rr := httptest.NewRecorder()
	h.Withdraw(rr, authedBal(http.MethodPost, "/api/user/balance/withdraw", `{"order":"2377225624","sum":729.98}`))
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "7", f.gotUserID)
	require.Equal(t, "2377225624", f.gotOrder)
	require.Equal(t, money.FromFloat(729.98), f.gotSum)
}

func TestWithdrawHandlerInsufficient(t *testing.T) {
	h := handlerWith(&fakeStore{withdrawErr: repository.ErrInsufficientFunds})
	rr := httptest.NewRecorder()
	h.Withdraw(rr, authedBal(http.MethodPost, "/api/user/balance/withdraw", `{"order":"2377225624","sum":10}`))
	require.Equal(t, http.StatusPaymentRequired, rr.Code)
}

func TestWithdrawHandlerBadLuhn(t *testing.T) {
	f := &fakeStore{}
	h := handlerWith(f)
	rr := httptest.NewRecorder()
	h.Withdraw(rr, authedBal(http.MethodPost, "/api/user/balance/withdraw", `{"order":"12345678902","sum":10}`))
	require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	require.Empty(t, f.gotOrder, "store must not be called on bad luhn")
}

func TestWithdrawHandlerNonPositiveSum(t *testing.T) {
	f := &fakeStore{}
	h := handlerWith(f)
	rr := httptest.NewRecorder()
	h.Withdraw(rr, authedBal(http.MethodPost, "/api/user/balance/withdraw", `{"order":"2377225624","sum":-5}`))
	require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	require.Empty(t, f.gotOrder, "store must not be called on non-positive sum")
}

func TestWithdrawHandlerUnauthorized(t *testing.T) {
	h := handlerWith(&fakeStore{})
	rr := httptest.NewRecorder()
	h.Withdraw(rr, httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", nil))
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestWithdrawHandlerBadJSON(t *testing.T) {
	h := handlerWith(&fakeStore{})
	rr := httptest.NewRecorder()
	h.Withdraw(rr, authedBal(http.MethodPost, "/api/user/balance/withdraw", `{bad`))
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestWithdrawHandlerInternal(t *testing.T) {
	h := handlerWith(&fakeStore{withdrawErr: context.DeadlineExceeded})
	rr := httptest.NewRecorder()
	h.Withdraw(rr, authedBal(http.MethodPost, "/api/user/balance/withdraw", `{"order":"2377225624","sum":10}`))
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}
