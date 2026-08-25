package model_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/superserj/gophermart/internal/model"
)

func TestAuthRequestUnmarshal(t *testing.T) {
	var req model.AuthRequest
	require.NoError(t, json.Unmarshal([]byte(`{"login":"bob","password":"pw"}`), &req))
	assert.Equal(t, "bob", req.Login)
	assert.Equal(t, "pw", req.Password)
}
