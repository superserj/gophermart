package repository

import (
	"context"
	"time"

	"github.com/superserj/gophermart/internal/luhn"
	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
)

// Withdraw списывает баллы в счёт заказа в транзакции, сериализуя конкурентные
// списания одного пользователя через FOR UPDATE. Возвращает ErrInsufficientFunds
// при нехватке средств, ErrInvalidWithdrawal при невалидном номере/неположительной сумме.
func (s *DBStorage) Withdraw(ctx context.Context, userID int64, order string, sum money.Points) error {
	if sum <= 0 || !luhn.Valid(order) {
		return ErrInvalidWithdrawal
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var locked int64
	if err := tx.QueryRow(ctx, `SELECT id FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&locked); err != nil {
		return err
	}
	var current int64
	if err := tx.QueryRow(ctx,
		`SELECT
		   COALESCE((SELECT SUM(accrual) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'), 0)
		     - COALESCE((SELECT SUM(sum) FROM withdrawals WHERE user_id = $1), 0)`,
		userID).Scan(&current); err != nil {
		return err
	}
	if money.Points(current) < sum {
		return ErrInsufficientFunds
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`,
		userID, order, int64(sum)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ListWithdrawals возвращает списания пользователя от новых к старым.
func (s *DBStorage) ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT order_number, sum, processed_at FROM withdrawals
		 WHERE user_id = $1 ORDER BY processed_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Withdrawal
	for rows.Next() {
		var (
			order string
			sum   int64
			ts    time.Time
		)
		if err := rows.Scan(&order, &sum, &ts); err != nil {
			return nil, err
		}
		result = append(result, model.Withdrawal{Order: order, Sum: money.Points(sum), ProcessedAt: ts})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
