package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
)

func seedProcessed(t *testing.T, st *DBStorage, uid int64, number string, accrual money.Points) {
	t.Helper()
	_, err := st.SaveOrder(context.Background(), number, uid)
	require.NoError(t, err)
	require.NoError(t, st.ApplyAccrual(context.Background(), number, model.StatusProcessed, accrual))
}

func TestGetBalanceAccrualOnly(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	uid := mustCreateUser(t, st, "u1")

	seedProcessed(t, st, uid, "9278923470", money.FromFloat(729.98))
	seedProcessed(t, st, uid, "12345678903", money.FromFloat(500))

	current, withdrawn, err := st.GetBalance(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, money.FromFloat(1229.98), current)
	require.Equal(t, money.Points(0), withdrawn)
}

func TestGetBalanceEmpty(t *testing.T) {
	st := newTestStore(t)
	uid := mustCreateUser(t, st, "u1")
	current, withdrawn, err := st.GetBalance(context.Background(), uid)
	require.NoError(t, err)
	require.Equal(t, money.Points(0), current)
	require.Equal(t, money.Points(0), withdrawn)
}
