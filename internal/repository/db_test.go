package repository

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URI")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URI not set; skipping live DB ping")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	s, err := NewDBStorage(ctx, dsn)
	if err != nil {
		t.Fatalf("NewDBStorage: %v", err)
	}
	defer s.Close()
	if err := s.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}
