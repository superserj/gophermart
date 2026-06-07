package accrual

import "github.com/superserj/gophermart/internal/model"

// mapStatus переводит статус accrual в наш доменный статус.
func mapStatus(accrualStatus string) string {
	switch accrualStatus {
	case "INVALID":
		return model.StatusInvalid
	case "PROCESSED":
		return model.StatusProcessed
	default: // REGISTERED, PROCESSING и неизвестное — продолжаем опрос
		return model.StatusProcessing
	}
}
