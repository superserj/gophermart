package repository

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// CreateUser вставляет пользователя; занятый логин → ErrLoginTaken. Числовой
// первичный ключ конвертируется в строковый идентификатор на границе слоя.
func (s *DBStorage) CreateUser(ctx context.Context, login, passwordHash string) (string, error) {
	var id int64
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1, $2)
		 ON CONFLICT (login) DO NOTHING
		 RETURNING id`,
		login, passwordHash).Scan(&id)
	if err == nil {
		return strconv.FormatInt(id, 10), nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrLoginTaken
	}
	return "", err
}

// GetUserByLogin возвращает строковый id и bcrypt-хэш; нет строки → ErrUserNotFound.
func (s *DBStorage) GetUserByLogin(ctx context.Context, login string) (string, string, error) {
	var (
		id   int64
		hash string
	)
	err := s.pool.QueryRow(ctx,
		`SELECT id, password_hash FROM users WHERE login = $1`, login).Scan(&id, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrUserNotFound
	}
	if err != nil {
		return "", "", err
	}
	return strconv.FormatInt(id, 10), hash, nil
}
