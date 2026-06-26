package repository

import (
	"context"
	"strconv"

	"github.com/superserj/gophermart/internal/money"
)

// GetBalance деривирует баланс агрегатом: начислено по PROCESSED минус списано.
func (s *DBStorage) GetBalance(ctx context.Context, userID string) (money.Points, money.Points, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	var current, withdrawn int64
	err = s.pool.QueryRow(ctx,
		`SELECT
		   COALESCE((SELECT SUM(accrual) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'), 0)
		     - COALESCE((SELECT SUM(sum) FROM withdrawals WHERE user_id = $1), 0),
		   COALESCE((SELECT SUM(sum) FROM withdrawals WHERE user_id = $1), 0)`,
		uid).Scan(&current, &withdrawn)
	if err != nil {
		return 0, 0, err
	}
	return money.Points(current), money.Points(withdrawn), nil
}
