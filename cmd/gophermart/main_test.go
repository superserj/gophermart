package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubPinger struct{ err error }

func (s stubPinger) Ping(context.Context) error { return s.err }

func TestPingRouteOK(t *testing.T) {
	rec := httptest.NewRecorder()
	pingHandler(stubPinger{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET /ping = %d, want 200", res.StatusCode)
	}
}

func TestPingRouteDBDown(t *testing.T) {
	rec := httptest.NewRecorder()
	pingHandler(stubPinger{err: context.DeadlineExceeded}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	res := rec.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("GET /ping (db down) = %d, want 500", res.StatusCode)
	}
}
