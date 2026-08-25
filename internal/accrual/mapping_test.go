package accrual

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/superserj/gophermart/internal/model"
)

func TestMapStatus(t *testing.T) {
	cases := map[string]string{
		"REGISTERED": model.StatusProcessing,
		"PROCESSING": model.StatusProcessing,
		"INVALID":    model.StatusInvalid,
		"PROCESSED":  model.StatusProcessed,
		"WHATEVER":   model.StatusProcessing,
	}
	for in, want := range cases {
		assert.Equal(t, want, mapStatus(in), in)
	}
}
