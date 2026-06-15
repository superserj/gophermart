package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInitialize(t *testing.T) {
	if err := Initialize("info"); err != nil {
		t.Fatalf("valid level: unexpected error %v", err)
	}
	if Log == nil {
		t.Fatal("Log must be set after Initialize")
	}
	if err := Initialize("nonsense"); err == nil {
		t.Fatal("invalid level must return error")
	}
}

func TestWithLogging(t *testing.T) {
	if err := Initialize("info"); err != nil {
		t.Fatalf("initialize: %v", err)
	}
	h := WithLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
