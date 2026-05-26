// Package model содержит доменные сущности и DTO Гофермарта.
package model

import (
	"time"

	"github.com/superserj/gophermart/internal/money"
)

// Статусы обработки заказа. INVALID и PROCESSED — терминальные.
const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

// User — пользователь системы.
type User struct {
	ID           int64
	Login        string
	PasswordHash string
}

// Order — заказ для выдачи в GET /api/user/orders.
// Accrual опускается из JSON при нулевом начислении (NEW/PROCESSING/INVALID).
type Order struct {
	Number     string       `json:"number"`
	Status     string       `json:"status"`
	Accrual    money.Points `json:"accrual,omitempty"`
	UploadedAt time.Time    `json:"uploaded_at"`
}

// Balance — баланс пользователя.
type Balance struct {
	Current   money.Points `json:"current"`
	Withdrawn money.Points `json:"withdrawn"`
}

// Withdrawal — одно списание баллов.
type Withdrawal struct {
	Order       string       `json:"order"`
	Sum         money.Points `json:"sum"`
	ProcessedAt time.Time    `json:"processed_at"`
}
