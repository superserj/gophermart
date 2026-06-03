package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superserj/gophermart/internal/model"
)

func TestSaveOrder(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
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
	ctx := context.Background()
	u := mustCreateUser(t, st, "u1")

	_, err := st.SaveOrder(ctx, "9278923470", u)
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)
	_, err = st.SaveOrder(ctx, "12345678903", u)
	require.NoError(t, err)

	orders, err := st.ListOrdersByUser(ctx, u)
	require.NoError(t, err)
	require.Len(t, orders, 2)
	assert.Equal(t, "12345678903", orders[0].Number) // newest first
	assert.Equal(t, model.StatusNew, orders[0].Status)

	none, err := st.ListOrdersByUser(ctx, 99999)
	require.NoError(t, err)
	assert.Empty(t, none)
}
