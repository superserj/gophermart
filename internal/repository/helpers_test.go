package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *DBStorage {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URI")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URI not set")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	st, err := NewDBStorage(ctx, dsn)
	require.NoError(t, err)
	_, err = st.pool.Exec(t.Context(), `TRUNCATE withdrawals, orders, users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func mustCreateUser(t *testing.T, st *DBStorage, login string) string {
	t.Helper()
	id, err := st.CreateUser(t.Context(), login, "hash")
	require.NoError(t, err)
	return id
}
