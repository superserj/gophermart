package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShouldCompressTextPlain(t *testing.T) {
	if !shouldCompress("text/plain; charset=utf-8") {
		t.Fatal("text/plain must be compressible")
	}
	if !shouldCompress("application/json") {
		t.Fatal("application/json must be compressible")
	}
	if shouldCompress("image/png") {
		t.Fatal("image/png must not be compressed")
	}
}

func TestGzipResponseTextPlain(t *testing.T) {
	h := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("12345678903"))
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer zr.Close()
	body, _ := io.ReadAll(zr)
	if !bytes.Equal(body, []byte("12345678903")) {
		t.Fatalf("decompressed body = %q", body)
	}
}

func TestGzipRequestDecompression(t *testing.T) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write([]byte("compressed-payload")); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}

	var got []byte
	h := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		got = body
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !bytes.Equal(got, []byte("compressed-payload")) {
		t.Fatalf("decompressed request body = %q, want compressed-payload", got)
	}
}
