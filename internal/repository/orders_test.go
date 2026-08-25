package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superserj/gophermart/internal/model"
)

func TestSaveOrder(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()
	u1 := mustCreateUser(t, st, "u1")
	u2 := mustCreateUser(t, st, "u2")

	existed, err := st.SaveOrder(ctx, "12345678903", u1)
	require.NoError(t, err)
	assert.False(t, existed)

	existed, err = st.SaveOrder(ctx, "12345678903", u1)
	require.NoError(t, err)
	assert.True(t, existed)

	_, err = st.SaveOrder(ctx, "12345678903", u2)
	assert.ErrorIs(t, err, ErrOrderOwnedByOther)
}

func TestListOrdersByUser(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()
	u := mustCreateUser(t, st, "u1")

	// Явные uploaded_at делают порядок сортировки детерминированным и не зависят
	// от точности системного таймера между двумя вставками (флак на медленном CI).
	older := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)
	_, err := st.pool.Exec(ctx,
		`INSERT INTO orders (number, user_id, uploaded_at) VALUES ($1, $2::bigint, $3), ($4, $2::bigint, $5)`,
		"9278923470", u, older, "12345678903", newer)
	require.NoError(t, err)

	orders, err := st.ListOrdersByUser(ctx, u)
	require.NoError(t, err)
	require.Len(t, orders, 2)
	assert.Equal(t, "12345678903", orders[0].Number) // newest first
	assert.Equal(t, model.StatusNew, orders[0].Status)

	none, err := st.ListOrdersByUser(ctx, "99999")
	require.NoError(t, err)
	assert.Empty(t, none)
}
