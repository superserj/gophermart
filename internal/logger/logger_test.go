package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	log, err := New("info")
	if err != nil {
		t.Fatalf("valid level: unexpected error %v", err)
	}
	if log == nil {
		t.Fatal("New must return a logger")
	}
	if _, err := New("nonsense"); err == nil {
		t.Fatal("invalid level must return error")
	}
}

func TestWithLogging(t *testing.T) {
	h := WithLogging(zap.NewNop())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("body = %q, want hello", rec.Body.String())
	}
}
