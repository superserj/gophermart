package model_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/superserj/gophermart/internal/model"
)

func TestWithdrawRequestDecode(t *testing.T) {
	var req model.WithdrawRequest
	require.NoError(t, json.Unmarshal([]byte(`{"order":"2377225624","sum":729.98}`), &req))
	require.Equal(t, "2377225624", req.Order)
	require.Equal(t, 729.98, req.Sum)
}
