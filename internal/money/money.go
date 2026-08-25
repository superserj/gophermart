// Package money хранит денежные суммы как целые баллы×100 (копейки),
// исключая дрейф float при хранении и агрегации.
package money

import (
	"math"
	"strconv"
)

// Points — сумма в баллах×100.
type Points int64

// FromFloat переводит баллы (float) в Points с округлением до копейки.
// NaN/Inf (мусор от accrual) дают implementation-defined int64 при конверсии,
// поэтому отбраковываются в 0.
func FromFloat(f float64) Points {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return Points(0)
	}
	return Points(math.Round(f * 100))
}

// Float возвращает сумму в баллах.
func (p Points) Float() float64 { return float64(p) / 100 }

// MarshalJSON выводит число с двумя знаками после точки: 72998 → 729.98.
func (p Points) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(p.Float(), 'f', 2, 64)), nil
}

// UnmarshalJSON читает JSON-число (баллы) и округляет до копейки.
func (p *Points) UnmarshalJSON(data []byte) error {
	f, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return err
	}
	*p = FromFloat(f)
	return nil
}
