package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
)

// SaveOrder регистрирует номер заказа за пользователем.
// existed=false → создан (202); existed=true → уже свой (200); ErrOrderOwnedByOther → 409.
func (s *DBStorage) SaveOrder(ctx context.Context, number string, userID int64) (bool, error) {
	var inserted string
	err := s.pool.QueryRow(ctx,
		`INSERT INTO orders (number, user_id) VALUES ($1, $2)
		 ON CONFLICT (number) DO NOTHING
		 RETURNING number`,
		number, userID).Scan(&inserted)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}
	var owner int64
	if err := s.pool.QueryRow(ctx,
		`SELECT user_id FROM orders WHERE number = $1`, number).Scan(&owner); err != nil {
		return false, err
	}
	if owner == userID {
		return true, nil
	}
	return false, ErrOrderOwnedByOther
}

// ListOrdersByUser возвращает заказы пользователя от новых к старым.
func (s *DBStorage) ListOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT number, status, accrual, uploaded_at
		 FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.Order
	for rows.Next() {
		var (
			o       model.Order
			accrual int64
		)
		if err := rows.Scan(&o.Number, &o.Status, &accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		o.Accrual = money.Points(accrual)
		result = append(result, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
