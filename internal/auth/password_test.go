package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("s3cret")
	require.NoError(t, err)
	assert.NotEqual(t, "s3cret", hash)
	assert.NoError(t, CheckPassword(hash, "s3cret"))
	assert.Error(t, CheckPassword(hash, "wrong"))
}

func TestHashPasswordUniqueSalt(t *testing.T) {
	h1, err := HashPassword("same")
	require.NoError(t, err)
	h2, err := HashPassword("same")
	require.NoError(t, err)
	assert.NotEqual(t, h1, h2, "bcrypt must use a random salt")
}
