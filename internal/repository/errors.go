package repository

import "errors"

// Типизированные ошибки доменного слоя репозитория.
var (
	// ErrLoginTaken — логин уже занят другим пользователем.
	ErrLoginTaken = errors.New("login already taken")
	// ErrUserNotFound — пользователь с таким логином не найден.
	ErrUserNotFound = errors.New("user not found")
	// ErrOrderOwnedByOther — заказ уже зарегистрирован другим пользователем.
	ErrOrderOwnedByOther = errors.New("order owned by another user")
	// ErrInsufficientFunds — на балансе недостаточно баллов для списания.
	ErrInsufficientFunds = errors.New("insufficient funds")
	// ErrInvalidWithdrawal — невалидный номер заказа или неположительная сумма списания.
	ErrInvalidWithdrawal = errors.New("invalid withdrawal")
)
