package repository

import (
	"context"

	"github.com/superserj/gophermart/internal/money"
)

// ListUnfinishedOrders возвращает номера заказов в незавершённых статусах.
func (s *DBStorage) ListUnfinishedOrders(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT number FROM orders WHERE status IN ('NEW', 'PROCESSING')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var numbers []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		numbers = append(numbers, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return numbers, nil
}

// ApplyAccrual идемпотентно обновляет статус и начисление заказа,
// не трогая уже терминальные заказы.
func (s *DBStorage) ApplyAccrual(ctx context.Context, number, status string, accrual money.Points) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE orders SET status = $2, accrual = $3
		 WHERE number = $1 AND status NOT IN ('INVALID', 'PROCESSED')`,
		number, status, int64(accrual))
	return err
}
