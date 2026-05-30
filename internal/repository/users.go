package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// CreateUser вставляет пользователя; занятый логин → ErrLoginTaken.
func (s *DBStorage) CreateUser(ctx context.Context, login, passwordHash string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1, $2)
		 ON CONFLICT (login) DO NOTHING
		 RETURNING id`,
		login, passwordHash).Scan(&id)
	if err == nil {
		return id, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrLoginTaken
	}
	return 0, err
}

// GetUserByLogin возвращает id и bcrypt-хэш; нет строки → ErrUserNotFound.
func (s *DBStorage) GetUserByLogin(ctx context.Context, login string) (int64, string, error) {
	var (
		id   int64
		hash string
	)
	err := s.pool.QueryRow(ctx,
		`SELECT id, password_hash FROM users WHERE login = $1`, login).Scan(&id, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, "", ErrUserNotFound
	}
	if err != nil {
		return 0, "", err
	}
	return id, hash, nil
}
