package repository

import (
	"context"
	"errors"
	"sync"
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

func TestWithdrawSuccessAndList(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	uid := mustCreateUser(t, st, "u1")
	seedProcessed(t, st, uid, "9278923470", money.FromFloat(1000))

	require.NoError(t, st.Withdraw(ctx, uid, "2377225624", money.FromFloat(300)))
	current, withdrawn, err := st.GetBalance(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, money.FromFloat(700), current)
	require.Equal(t, money.FromFloat(300), withdrawn)

	list, err := st.ListWithdrawals(ctx, uid)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "2377225624", list[0].Order)
	require.Equal(t, money.FromFloat(300), list[0].Sum)
	require.False(t, list[0].ProcessedAt.IsZero())
}

func TestWithdrawInsufficient(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	uid := mustCreateUser(t, st, "u1")
	seedProcessed(t, st, uid, "9278923470", money.FromFloat(100))
	err := st.Withdraw(ctx, uid, "2377225624", money.FromFloat(300))
	require.ErrorIs(t, err, ErrInsufficientFunds)
	list, err := st.ListWithdrawals(ctx, uid)
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestWithdrawConcurrentSerialization(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	uid := mustCreateUser(t, st, "u1")
	seedProcessed(t, st, uid, "9278923470", money.FromFloat(100))

	orders := []string{"2377225624", "12345678903"}
	results := make([]error, len(orders))
	var wg sync.WaitGroup
	for i, ord := range orders {
		wg.Add(1)
		go func(i int, ord string) {
			defer wg.Done()
			results[i] = st.Withdraw(ctx, uid, ord, money.FromFloat(60))
		}(i, ord)
	}
	wg.Wait()

	var ok, insuff int
	for _, err := range results {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, ErrInsufficientFunds):
			insuff++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	require.Equal(t, 1, ok)
	require.Equal(t, 1, insuff)
	current, _, err := st.GetBalance(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, money.FromFloat(40), current)
}
