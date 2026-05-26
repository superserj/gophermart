package model_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/superserj/gophermart/internal/model"
	"github.com/superserj/gophermart/internal/money"
)

func TestOrderJSON(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	ts := time.Date(2020, 12, 10, 15, 15, 45, 0, loc)

	processed := model.Order{Number: "9278923470", Status: model.StatusProcessed, Accrual: 72998, UploadedAt: ts}
	b, _ := json.Marshal(processed)
	s := string(b)
	if !strings.Contains(s, `"accrual":729.98`) {
		t.Fatalf("accrual must be present: %s", s)
	}
	if !strings.Contains(s, `"uploaded_at":"2020-12-10T15:15:45+03:00"`) {
		t.Fatalf("uploaded_at must be RFC3339: %s", s)
	}

	pending := model.Order{Number: "9278923470", Status: model.StatusNew, UploadedAt: ts}
	b2, _ := json.Marshal(pending)
	if strings.Contains(string(b2), "accrual") {
		t.Fatalf("accrual must be omitted when zero: %s", b2)
	}
}

func TestBalanceJSON(t *testing.T) {
	b, _ := json.Marshal(model.Balance{Current: money.FromFloat(729.98), Withdrawn: 0})
	if string(b) != `{"current":729.98,"withdrawn":0.00}` {
		t.Fatalf("balance json = %s", b)
	}
}

func TestWithdrawalJSON(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	w := model.Withdrawal{Order: "2377225624", Sum: money.FromFloat(500), ProcessedAt: time.Date(2020, 12, 9, 16, 9, 57, 0, loc)}
	b, _ := json.Marshal(w)
	s := string(b)
	if !strings.Contains(s, `"sum":500.00`) || !strings.Contains(s, `"order":"2377225624"`) {
		t.Fatalf("withdrawal json = %s", s)
	}
	if !strings.Contains(s, `"processed_at":"2020-12-09T16:09:57+03:00"`) {
		t.Fatalf("processed_at must be RFC3339: %s", s)
	}
}
