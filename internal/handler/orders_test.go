package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superserj/gophermart/internal/auth"
	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
	"github.com/superserj/gophermart/internal/repository"
)

func authedReq(method, body, userID string) *http.Request {
	req := httptest.NewRequest(method, "/api/user/orders", strings.NewReader(body))
	if userID != "" {
		req = req.WithContext(auth.WithUserID(req.Context(), userID))
	}
	return req
}

func TestUploadOrder(t *testing.T) {
	t.Run("new 202", func(t *testing.T) {
		f := &fakeStore{saveExisted: false}
		h := &Handler{store: f}
		rr := httptest.NewRecorder()
		h.UploadOrder(rr, authedReq(http.MethodPost, "12345678903", "1"))
		assert.Equal(t, http.StatusAccepted, rr.Code)
		assert.Equal(t, "12345678903", f.gotNumber)
		assert.Equal(t, int64(1), f.gotUserID)
	})
	t.Run("already own 200", func(t *testing.T) {
		h := &Handler{store: &fakeStore{saveExisted: true}}
		rr := httptest.NewRecorder()
		h.UploadOrder(rr, authedReq(http.MethodPost, "12345678903", "1"))
		assert.Equal(t, http.StatusOK, rr.Code)
	})
	t.Run("owned by other 409", func(t *testing.T) {
		h := &Handler{store: &fakeStore{saveErr: repository.ErrOrderOwnedByOther}}
		rr := httptest.NewRecorder()
		h.UploadOrder(rr, authedReq(http.MethodPost, "12345678903", "1"))
		assert.Equal(t, http.StatusConflict, rr.Code)
	})
	t.Run("luhn invalid 422 no store", func(t *testing.T) {
		f := &fakeStore{}
		h := &Handler{store: f}
		rr := httptest.NewRecorder()
		h.UploadOrder(rr, authedReq(http.MethodPost, "12345678902", "1"))
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		assert.False(t, f.saveCalled)
	})
	t.Run("no auth 401", func(t *testing.T) {
		h := &Handler{store: &fakeStore{}}
		rr := httptest.NewRecorder()
		h.UploadOrder(rr, authedReq(http.MethodPost, "12345678903", ""))
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
	t.Run("empty body 400", func(t *testing.T) {
		h := &Handler{store: &fakeStore{}}
		rr := httptest.NewRecorder()
		h.UploadOrder(rr, authedReq(http.MethodPost, "   ", "1"))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
	t.Run("oversized body 400 no store", func(t *testing.T) {
		f := &fakeStore{}
		h := &Handler{store: f}
		rr := httptest.NewRecorder()
		huge := strings.Repeat("0", 100*1024)
		h.UploadOrder(rr, authedReq(http.MethodPost, huge, "1"))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.False(t, f.saveCalled)
	})
}

func TestListOrders(t *testing.T) {
	t.Run("has orders 200 json", func(t *testing.T) {
		f := &fakeStore{orders: []model.Order{{Number: "9278923470", Status: model.StatusProcessed, Accrual: money.FromFloat(729.98), UploadedAt: time.Now()}}}
		h := &Handler{store: f}
		rr := httptest.NewRecorder()
		h.ListOrders(rr, authedReq(http.MethodGet, "", "1"))
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		var got []model.Order
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
		require.Len(t, got, 1)
		assert.Equal(t, money.FromFloat(729.98), got[0].Accrual)
	})
	t.Run("empty 204 json", func(t *testing.T) {
		h := &Handler{store: &fakeStore{orders: nil}}
		rr := httptest.NewRecorder()
		h.ListOrders(rr, authedReq(http.MethodGet, "", "1"))
		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	})
	t.Run("no auth 401", func(t *testing.T) {
		h := &Handler{store: &fakeStore{}}
		rr := httptest.NewRecorder()
		h.ListOrders(rr, authedReq(http.MethodGet, "", ""))
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
