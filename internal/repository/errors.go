package repository

import "errors"

// Типизированные ошибки доменного слоя репозитория.
var (
	ErrLoginTaken        = errors.New("login already taken")
	ErrUserNotFound      = errors.New("user not found")
	ErrOrderOwnedByOther = errors.New("order owned by another user")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidWithdrawal = errors.New("invalid withdrawal")
)
