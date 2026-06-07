package accrual

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientGetOrderOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orders/79927398713", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"order":"79927398713","status":"PROCESSED","accrual":729.98}`))
	}))
	defer srv.Close()
	info, err := NewClient(srv.URL).GetOrder(context.Background(), "79927398713")
	require.NoError(t, err)
	assert.Equal(t, "PROCESSED", info.Status)
	assert.InDelta(t, 729.98, info.Accrual, 0.001)
}

func TestClientNoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer srv.Close()
	_, err := NewClient(srv.URL).GetOrder(context.Background(), "1")
	assert.ErrorIs(t, err, ErrNoContent)
}

func TestClientThrottled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "3")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	_, err := NewClient(srv.URL).GetOrder(context.Background(), "1")
	var tm *TooManyRequestsError
	require.ErrorAs(t, err, &tm)
	assert.Equal(t, 3*time.Second, tm.RetryAfter)
}

func TestClientThrottledNoHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTooManyRequests) }))
	defer srv.Close()
	_, err := NewClient(srv.URL).GetOrder(context.Background(), "1")
	var tm *TooManyRequestsError
	require.ErrorAs(t, err, &tm)
	assert.Equal(t, time.Second, tm.RetryAfter)
}

func TestClientServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
	defer srv.Close()
	_, err := NewClient(srv.URL).GetOrder(context.Background(), "1")
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrNoContent)
}
