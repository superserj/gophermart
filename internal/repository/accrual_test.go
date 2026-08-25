package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
)

func TestApplyAccrualIdempotent(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()
	u := mustCreateUser(t, st, "u1")
	const number = "79927398713"
	_, err := st.SaveOrder(ctx, number, u)
	require.NoError(t, err)

	require.NoError(t, st.ApplyAccrual(ctx, number, model.StatusProcessed, money.FromFloat(729.98)))

	unfinished, err := st.ListUnfinishedOrders(ctx)
	require.NoError(t, err)
	assert.NotContains(t, unfinished, number)

	// повторный apply не перетирает терминальный статус/начисление
	require.NoError(t, st.ApplyAccrual(ctx, number, model.StatusProcessing, money.FromFloat(1)))
	orders, err := st.ListOrdersByUser(ctx, u)
	require.NoError(t, err)
	require.Len(t, orders, 1)
	assert.Equal(t, model.StatusProcessed, orders[0].Status)
	assert.Equal(t, money.FromFloat(729.98), orders[0].Accrual)
}

func TestListUnfinishedOrders(t *testing.T) {
	st := newTestStore(t)
	ctx := t.Context()
	u := mustCreateUser(t, st, "u1")
	_, err := st.SaveOrder(ctx, "79927398713", u)
	require.NoError(t, err)
	got, err := st.ListUnfinishedOrders(ctx)
	require.NoError(t, err)
	assert.Contains(t, got, "79927398713")
}
